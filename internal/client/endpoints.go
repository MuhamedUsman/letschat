package client

const (
	baseUrl               = "http://localhost:8080/v1"
	usersEndpoint         = "/users"
	tokensEndpoint        = "/tokens"
	messagesEndpoint      = "/recvMsgs"
	conversationsEndpoint = "/conversations"
	wsBaseUrl             = "ws://localhost:8080"
	websocketsEndpoint    = "/sub"

	registerUser         = baseUrl + usersEndpoint // POST
	getByUniqueField     = baseUrl + usersEndpoint // GET
	getCurrentActiveUser = getByUniqueField + "/current"
	searchUser           = getByUniqueField
	updateUser           = baseUrl + usersEndpoint               // PUT
	activateUser         = baseUrl + usersEndpoint + "/activate" // POST

	generateOTP  = baseUrl + tokensEndpoint + "/otp"  // POST
	authenticate = baseUrl + tokensEndpoint + "/auth" // POST

	getConversations = baseUrl + conversationsEndpoint

	getMessages = baseUrl + messagesEndpoint // GET

	subscribeTo = wsBaseUrl + websocketsEndpoint
)
