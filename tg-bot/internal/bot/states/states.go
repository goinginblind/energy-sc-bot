package states

const (
	StateStart                  = ""
	StateAwaitingLoginInput     = "awaiting_login_input"
	StateAwaitingOTP            = "awaiting_otp"
	StateLoggedIn               = "logged_in"
	StateGeneralInquiry         = "general_inquiry"
	StateAwaitingAgentIssuePost = "awaiting_agent_issue_post"
	StateAgentChat              = "agent_chat"
)
