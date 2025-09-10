package audio

// MockRecorder interface implementation
type MockRecorder struct{}

func NewMockRecorder() *MockRecorder {
return &MockRecorder{}
}

func (r *MockRecorder) Start(ctx interface{}, options interface{}, callback interface{}) error {
return nil
}

func (r *MockRecorder) Stop() (interface{}, error) {
return nil, nil
}

func (r *MockRecorder) Pause() error {
return nil
}

func (r *MockRecorder) Resume() error {
return nil
}

func (r *MockRecorder) GetState() RecordingState {
return RecordingStateIdle
}

func (r *MockRecorder) GetLevel() float32 {
return 0.0
}

func (r *MockRecorder) GetDuration() interface{} {
return nil
}

func (r *MockRecorder) Cancel() error {
return nil
}

func (r *MockRecorder) IsSupported() bool {
return true
}
