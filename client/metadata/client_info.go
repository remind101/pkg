package metadata

// ClientInfo holds metadata about the client.
type ClientInfo struct {
	ServiceName string // Name of the service. For example: ExampleOrg
	Endpoint    string // Base URL for the client. For example: http://api.example.org
}
