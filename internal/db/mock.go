package db

type MockService struct{}

func (m *MockService) Health() map[string]string {
	return map[string]string{
		"status":  "up",
		"message": "Mock database is healthy",
	}
}

func (m *MockService) Close() error {
	return nil
}

func (m *MockService) Queries() *Queries {
	return &Queries{} // Return empty Queries struct for documentation purposes
}
