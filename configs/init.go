package configs

// Init contains config of the Init command
type Init struct {
	WebSiteURL             string `json:"websiteurl"`
	DownloadBaseURL                string `json:"downloadbaseurl"`
	TimeInterval			int `json:"timeinterval"`
	VCenterIP                  string `json:"vcenterip"`
	VCenterUserName          string `json:"vcenterusername"`
	VCenterUserPwd                string `json:"vcenteruserpwd"`

}

// NewInit returns a new Init instance.
func NewInit() *Init {
	return &Init{}
}
       