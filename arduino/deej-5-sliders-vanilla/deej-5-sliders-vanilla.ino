const int NUM_SLIDERS = 4;
const int analogInputs[NUM_SLIDERS] = {A0, A1, A2, A3};
// BN GY WT YL

const int NUM_BUTTONS = 12;
const int buttonInputs[NUM_BUTTONS] = {2,3,4,5,6,7,8,9,10,11,12,13};

const int NUM_KNOBS = 2;
const int knobInputs[NUM_BUTTONS] = {A4, A5};

int analogSliderValues[NUM_SLIDERS];
int buttonValues[NUM_BUTTONS];
int knobValues[NUM_KNOBS];

void setup() { 
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  for (int i = 0; i < NUM_KNOBS; i++) {
    pinMode(knobInputs[i], INPUT);
  }

  for (int i = 0; i < NUM_BUTTONS; i++) {
    pinMode(buttonInputs[i], INPUT_PULLUP);
  }

  Serial.begin(9600);
}

void loop() {
  updateValues();
  sendValues(); // Actually send data (all the time)
  // printSliderValues(); // For debug
  delay(10);
}

void updateValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
     analogSliderValues[i] = analogRead(analogInputs[i]);
  }

  for (int i = 0; i < NUM_BUTTONS; i++) {
    buttonValues[i] = digitalRead(buttonInputs[i]);
  }

  for (int i = 0; i < NUM_KNOBS; i++) {
     knobValues[i] = analogRead(knobInputs[i]);
  }

}

void sendValues() {
  String builtString = String("");

  for (int i = 0; i < NUM_SLIDERS; i++) {
    builtString += String((int)analogSliderValues[i]);

    if (i < NUM_SLIDERS - 1) {
      builtString += String("|");
    }
  }

  builtString += String("/");

  for (int i = 0; i < NUM_BUTTONS; i++) {
    builtString += String((int)buttonValues[i]);
    if (i < NUM_BUTTONS - 1) {
      builtString += String("|");
    }
  }

  builtString += String("/");

  for (int i = 0; i < NUM_KNOBS; i++) {
    builtString += String((int)knobValues[i]);

    if (i < NUM_KNOBS - 1) {
      builtString += String("|");
    }
  }
  
  Serial.println(builtString);
}

void printSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    String printedString = String("Slider #") + String(i + 1) + String(": ") + String(analogSliderValues[i]) + String(" mV");
    Serial.write(printedString.c_str());

    if (i < NUM_SLIDERS - 1) {
      Serial.write(" | ");
    } else {
      Serial.write("\n");
    }
  }
}