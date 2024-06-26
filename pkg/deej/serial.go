package deej

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	// "regexp"
	"strconv"
	"strings"
	"time"
	"math"
	"github.com/jacobsa/go-serial/serial"
	"go.uber.org/zap"

	"github.com/llukad/deej-plus/pkg/deej/util"
)

// SerialIO provides a deej-aware abstraction layer to managing serial I/O
type SerialIO struct {
	comPort  string
	baudRate uint

	deej   *Deej
	logger *zap.SugaredLogger

	stopChannel chan bool
	connected   bool
	connOptions serial.OpenOptions
	conn        io.ReadWriteCloser

	lastKnownNumSliders        int
	currentSliderPercentValues []float32

	sliderMoveConsumers []chan SliderMoveEvent
}

// SliderMoveEvent represents a single slider move captured by deej
type SliderMoveEvent struct {
	SliderID     int
	PercentValue float32
}

var buttonValues []bool
var brightnessValue int
var timerValue int

// var expectedLinePattern = regexp.MustCompile(`^\d{1,4}(\|\d{1,4})*\/\d{1,4}(\|\d{1,4})*\/\d{1,4}$`)

// NewSerialIO creates a SerialIO instance that uses the provided deej
// instance's connection info to establish communications with the arduino chip
func NewSerialIO(deej *Deej, logger *zap.SugaredLogger) (*SerialIO, error) {
	logger = logger.Named("serial")

	sio := &SerialIO{
		deej:                deej,
		logger:              logger,
		stopChannel:         make(chan bool),
		connected:           false,
		conn:                nil,
		sliderMoveConsumers: []chan SliderMoveEvent{},
	}

	logger.Debug("Created serial i/o instance")

	// respond to config changes
	sio.setupOnConfigReload()

	return sio, nil
}

// Start attempts to connect to our arduino chip
func (sio *SerialIO) Start() error {

	// don't allow multiple concurrent connections
	if sio.connected {
		sio.logger.Warn("Already connected, can't start another without closing first")
		return errors.New("serial: connection already active")
	}

	// set minimum read size according to platform (0 for windows, 1 for linux)
	// this prevents a rare bug on windows where serial reads get congested,
	// resulting in significant lag
	minimumReadSize := 0
	if util.Linux() {
		minimumReadSize = 1
	}

	sio.connOptions = serial.OpenOptions{
		PortName:        sio.deej.config.ConnectionInfo.COMPort,
		BaudRate:        uint(sio.deej.config.ConnectionInfo.BaudRate),
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: uint(minimumReadSize),
	}

	sio.logger.Debugw("Attempting serial connection",
		"comPort", sio.connOptions.PortName,
		"baudRate", sio.connOptions.BaudRate,
		"minReadSize", minimumReadSize)

	var err error
	sio.conn, err = serial.Open(sio.connOptions)
	if err != nil {

		// might need a user notification here, TBD
		sio.logger.Warnw("Failed to open serial connection", "error", err)
		return fmt.Errorf("open serial connection: %w", err)
	}

	namedLogger := sio.logger.Named(strings.ToLower(sio.connOptions.PortName))

	namedLogger.Infow("Connected", "conn", sio.conn)
	sio.connected = true

	// read lines or await a stop
	go func() {
		connReader := bufio.NewReader(sio.conn)
		lineChannel := sio.readLine(namedLogger, connReader)

		for {
			select {
			case <-sio.stopChannel:
				sio.close(namedLogger)
			case line := <-lineChannel:
				sio.handleLine(namedLogger, line)
			}
		}
	}()

	return nil
}

// Stop signals us to shut down our serial connection, if one is active
func (sio *SerialIO) Stop() {
	if sio.connected {
		sio.logger.Debug("Shutting down serial connection")
		sio.stopChannel <- true
	} else {
		sio.logger.Debug("Not currently connected, nothing to stop")
	}
}

// SubscribeToSliderMoveEvents returns an unbuffered channel that receives
// a sliderMoveEvent struct every time a slider moves
func (sio *SerialIO) SubscribeToSliderMoveEvents() chan SliderMoveEvent {
	ch := make(chan SliderMoveEvent)
	sio.sliderMoveConsumers = append(sio.sliderMoveConsumers, ch)

	return ch
}

func (sio *SerialIO) setupOnConfigReload() {
	configReloadedChannel := sio.deej.config.SubscribeToChanges()

	const stopDelay = 50 * time.Millisecond

	go func() {
		for {
			select {
			case <-configReloadedChannel:

				// make any config reload unset our slider number to ensure process volumes are being re-set
				// (the next read line will emit SliderMoveEvent instances for all sliders)\
				// this needs to happen after a small delay, because the session map will also re-acquire sessions
				// whenever the config file is reloaded, and we don't want it to receive these move events while the map
				// is still cleared. this is kind of ugly, but shouldn't cause any issues
				go func() {
					<-time.After(stopDelay)
					sio.lastKnownNumSliders = 0
				}()

				// if connection params have changed, attempt to stop and start the connection
				if sio.deej.config.ConnectionInfo.COMPort != sio.connOptions.PortName ||
					uint(sio.deej.config.ConnectionInfo.BaudRate) != sio.connOptions.BaudRate {

					sio.logger.Info("Detected change in connection parameters, attempting to renew connection")
					sio.Stop()

					// let the connection close
					<-time.After(stopDelay)

					if err := sio.Start(); err != nil {
						sio.logger.Warnw("Failed to renew connection after parameter change", "error", err)
					} else {
						sio.logger.Debug("Renewed connection successfully")
					}
				}
			}
		}
	}()
}

func (sio *SerialIO) close(logger *zap.SugaredLogger) {
	if err := sio.conn.Close(); err != nil {
		logger.Warnw("Failed to close serial connection", "error", err)
	} else {
		logger.Debug("Serial connection closed")
	}

	sio.conn = nil
	sio.connected = false
}

func (sio *SerialIO) readLine(logger *zap.SugaredLogger, reader *bufio.Reader) chan string {
	ch := make(chan string)

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {

				if sio.deej.Verbose() {
					logger.Warnw("Failed to read line from serial", "error", err, "line", line)
				}

				// just ignore the line, the read loop will stop after this
				return
			}

			if sio.deej.Verbose() {
				logger.Debugw("Read new line", "line", line)
			}

			// deliver the line to the channel
			ch <- line
		}
	}()

	return ch
}

func (sio *SerialIO) handleLine(logger *zap.SugaredLogger, line string) {

	// this function receives an unsanitized line which is guaranteed to end with LF,
	// but most lines will end with CRLF. it may also have garbage instead of
	// deej-formatted values, so we must check for that! just ignore bad ones
	// if !expectedLinePattern.MatchString(line) {
	// 	logger.Infow(line)
	// 	// logger.Warnw("Wrong Line!")
	// 	return
	// }

	// trim the suffix
	line = strings.TrimSuffix(line, "\r\n")

	//split line on "/"

	splitLineAll := strings.Split(line, "/")

	// split on pipe (|), this gives a slice of numerical strings between "0" and "1023"
	splitLine := strings.Split(splitLineAll[0], "|")
	numSliders := len(splitLine)

	// update our slider count, if needed - this will send slider move events for all
	if numSliders != sio.lastKnownNumSliders {
		logger.Infow("Detected sliders", "amount", numSliders)
		sio.lastKnownNumSliders = numSliders
		sio.currentSliderPercentValues = make([]float32, numSliders)

		// reset everything to be an impossible value to force the slider move event later
		for idx := range sio.currentSliderPercentValues {
			sio.currentSliderPercentValues[idx] = -1.0
		}
	}

	// for each slider:
	moveEvents := []SliderMoveEvent{}
	for sliderIdx, stringValue := range splitLine {

		// convert string values to integers ("1023" -> 1023)
		number, _ := strconv.Atoi(stringValue)

		// turns out the first line could come out dirty sometimes (i.e. "4558|925|41|643|220")
		// so let's check the first number for correctness just in case
		if sliderIdx == 0 && number > 1023 {
			sio.logger.Debugw("Got malformed line from serial, ignoring", "line", line)
			return
		}

		// map the value from raw to a "dirty" float between 0 and 1 (e.g. 0.15451...)
		dirtyFloat := float32(number) / 1023.0

		// normalize it to an actual volume scalar between 0.0 and 1.0 with 2 points of precision
		normalizedScalar := util.NormalizeScalar(dirtyFloat)

		// if sliders are inverted, take the complement of 1.0
		if sio.deej.config.InvertSliders {
			normalizedScalar = 1 - normalizedScalar
		}

		// check if it changes the desired state (could just be a jumpy raw slider value)
		if util.SignificantlyDifferent(sio.currentSliderPercentValues[sliderIdx], normalizedScalar, sio.deej.config.NoiseReductionLevel) {

			// if it does, update the saved value and create a move event
			sio.currentSliderPercentValues[sliderIdx] = normalizedScalar

			moveEvents = append(moveEvents, SliderMoveEvent{
				SliderID:     sliderIdx,
				PercentValue: normalizedScalar,
			})

			if sio.deej.Verbose() {
				logger.Debugw("Slider moved", "event", moveEvents[len(moveEvents)-1])
			}
		}
	}

	// deliver move events if there are any, towards all potential consumers
	if len(moveEvents) > 0 {
		for _, consumer := range sio.sliderMoveConsumers {
			for _, moveEvent := range moveEvents {
				consumer <- moveEvent
			}
		}
	}

	if len(splitLineAll) < 3{
		return
	}

	buttons := strings.Split(splitLineAll[1], "|")
	numButtons := len(buttons)

	buttonsOk := true

	for _, button := range buttons{
		if!((button == "1") || (button == "0")){
			logger.Debugw("Wrong button values! Ignoring...")
			buttonsOk = false
		}
	}

	if buttonsOk{
		boolButtons, err := buttonsToBool(buttons)
		if err != nil {
			logger.Debugw("Cant parse button values!")
		}else{
			if numButtons == len(buttonValues){
				activatedButtons := make([]bool, numButtons)
				deactivatedButtons := make([]bool, numButtons)
				for i, button := range boolButtons{
					if button && !buttonValues[i]{
						activatedButtons[i] = true
					}else{
						activatedButtons[i] = false
					}

					if !button && buttonValues[i]{
						deactivatedButtons[i] = true
					}else{
						deactivatedButtons[i] = false
					}
				}
				// HERE WE DEFINE WHAT THE BUTTONS DO

				if activatedButtons[0]{
					err = OpenWebsite(sio.deej.config.openWebsite)
					if err != nil {
						logger.Debugw("Cant open website!", err)
					}else{
						logger.Infow("Website opened!")
					}
				}


				if activatedButtons[2]{
					err = ExecuteInTerminal(sio.deej.config.terminalCommand)
					if err != nil {
						logger.Debugw("Cant run the command!", err)
					}else{
						logger.Infow("Command ran!")
					}
				}


				if activatedButtons[4]{
					err = ChangeEyeProtection(0, 4, brightnessValue)
					if err != nil {
						logger.Debugw("Cant set eye protection!", err)
					}else{
						logger.Infow("Eye Protection set!")
					}
				}
				if activatedButtons[5]{
					err = ChangeColorMode(0, 33, brightnessValue)
					if err != nil {
						logger.Debugw("Cant change color mode!", err)
					}else{
						logger.Infow("Color mode set to rec.709!")
					}
				}
				if activatedButtons[6]{
					err = ShutdownIn(timerValue)
					if err != nil {
						logger.Debugw("Cant schedule shutdown!", err)
					}else{
						logger.Infow("Shutdown scheduled!")
					}
				}
				if activatedButtons[7]{
					err = StartVPN(sio.deej.config.VPN.Svc, sio.deej.config.VPN.Conf)
					if err != nil {
						logger.Debugw("Cant start VPN!", err)
						sio.deej.notifier.Notify("Error!", "Can't start the VPN!")
					}else{
						logger.Infow("VPN started!")
						sio.deej.notifier.Notify("VPN", "VPN started!")
					}
				}
				if deactivatedButtons[7]{
					err = StopVPN(sio.deej.config.VPN.Svc)
					if err != nil {
						logger.Debugw("Cant stop VPN!", err)
						sio.deej.notifier.Notify("Error!", "Can't stop the VPN!")
					}else{
						logger.Infow("VPN stopped!")
						sio.deej.notifier.Notify("VPN", "VPN stopped!")
					}
				}
				if activatedButtons[8]{
					err = LaunchApp("spotify.exe")
					if err != nil {
						logger.Debugw("Cant open spotify!", err)
					}else{
						logger.Infow("Spotify opened!")
					}
				}
				if activatedButtons[9]{
					err = PrevTrack()
					if err != nil {
						logger.Debugw("Cant do previous track", err)
					}else{
						logger.Infow("Previous track activated!")
					}
				}
				if activatedButtons[10]{
					err = PlayPause()
					if err != nil {
						logger.Debugw("Cant do play/pause", err)
					}else{
						logger.Infow("Play/pause activated!")
					}
				}
				if activatedButtons[11]{
					err = NextTrack()
					if err != nil {
						logger.Debugw("Cant do next track", err)
					}else{
						logger.Infow("Next track activated!")
					}
				}

			}
			buttonValues = boolButtons
		}
	}

	knobs := strings.Split(splitLineAll[2], "|")
	// numKnobs := len(knobs)

	knobsOk := true

	for _, knob := range knobs{
		knobVal, err := strconv.Atoi(knob)
		if err != nil{
			logger.Warnw("Cant parse knob values!", err)
		}
		if ((knobVal > 1023) && (knobVal < 0)){
			logger.Debugw("Wrong knob values! Ignoring...")
			knobsOk = false
		}
	}

	if knobsOk{
		timeVal, _ := strconv.Atoi(knobs[0])
		timerValue = convertToTime(timeVal)

		val, _ := strconv.Atoi(knobs[1])
		brghtVal := convertToBrightness(val)

		if brghtVal != brightnessValue{
			err := ChangeBrightness(0, brghtVal)
			if err != nil{
				logger.Warnw("Failed to change brightness:", err)
			}else{
				brightnessValue = brghtVal
				logger.Info("Changed brightness to: ", brghtVal)
			}
			
		}
	}
}

func buttonsToBool(buttons []string) ([]bool, error){
	boolButtons := make([]bool, len(buttons))
	for i, button := range buttons{
		if button == "1" {
			boolButtons[i] = false
		}else if button == "0" {
			boolButtons[i] = true
		}else{
			return nil, fmt.Errorf("Cant convert button states")
		}
	}

	return boolButtons, nil
}

func convertToBrightness(num int) int {
        floatNum := float64(num) / 1024 * 100


        roundedNum := math.Round(floatNum/10)*10

        // Convert the rounded float back to an integer
        result := int(roundedNum)

        // If the result is less than 0, set it to 0
        if result < 0 {
                result = 0
        }

        // If the result is greater than 100, set it to 100
        if result > 100 {
                result = 100
        }

        return result
}

func convertToTime(num int) int {
	maxTime := 3600
	minTime := 0

	mappedValue := minTime + (num * (maxTime - minTime) / 1023)
	return mappedValue
}
