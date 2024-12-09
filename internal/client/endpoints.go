package client

const (
	baseUrl               = "https://muhammadusman.site/v1"
	usersEndpoint         = "/users"
	tokensEndpoint        = "/tokens"
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

	subscribeTo = wsBaseUrl + websocketsEndpoint
)
