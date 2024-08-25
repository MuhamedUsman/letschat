package client

const (
	baseUrl            = "http://localhost:8080/v1"
	usersEndpoint      = "/users"
	tokensEndpoint     = "/tokens"
	messagesEndpoint   = "/messages"
	websocketsEndpoint = "/sub"

	registerUser     = baseUrl + usersEndpoint               // POST
	getByUniqueField = baseUrl + usersEndpoint               // GET
	updateUser       = baseUrl + usersEndpoint               // PUT
	activateUser     = baseUrl + usersEndpoint + "/activate" // POST

	generateOTP  = baseUrl + tokensEndpoint + "/otp"  // POST
	authenticate = baseUrl + tokensEndpoint + "/auth" // POST

	getMessages = baseUrl + messagesEndpoint // GET

	subscribeTo = baseUrl + websocketsEndpoint
)
