package model

type AuthResponse struct {
	Session struct {
		Valid    bool   `json:"valid"`
		Totp     bool   `json:"totp"`
		Sid      string `json:"sid"`
		Csrf     string `json:"csrf"`
		Validity int    `json:"validity"`
		Message  string `json:"message"`
	} `json:"session"`
}

type VersionResponse struct {
	Version struct {
		Core struct {
			Local struct {
				Branch  string `json:"branch"`
				Version string `json:"version"`
				Hash    string `json:"hash"`
			} `json:"local"`
			Remote struct {
				Version string `json:"version"`
				Hash    string `json:"hash"`
			} `json:"remote"`
		} `json:"core"`
		Web struct {
			Local struct {
				Branch  string `json:"branch"`
				Version string `json:"version"`
				Hash    string `json:"hash"`
			} `json:"local"`
			Remote struct {
				Version string `json:"version"`
				Hash    string `json:"hash"`
			} `json:"remote"`
		} `json:"web"`
		Ftl struct {
			Local struct {
				Branch  string `json:"branch"`
				Version string `json:"version"`
				Hash    string `json:"hash"`
				Date    string `json:"date"`
			} `json:"local"`
			Remote struct {
				Version string `json:"version"`
				Hash    string `json:"hash"`
			} `json:"remote"`
		} `json:"ftl"`
		Docker struct {
			Local  string `json:"local"`
			Remote string `json:"remote"`
		} `json:"docker"`
	} `json:"version"`
	Took float64 `json:"took"`
}

type ConfigResponse struct {
	Config map[string]interface{} `json:"config"`
}
