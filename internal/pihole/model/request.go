package model

type AuthRequest struct {
	Password string `json:"password"`
}

type PostGravityRequest struct {
	Group             bool `json:"group"`
	Adlist            bool `json:"adlist"`
	AdlistByGroup     bool `json:"adlist_by_group"`
	Domainlist        bool `json:"domainlist"`
	DomainlistByGroup bool `json:"domainlist_by_group"`
	Client            bool `json:"client"`
	ClientByGroup     bool `json:"client_by_group"`
}

type PostTeleporterRequest struct {
	Config     bool               `json:"config"`
	DHCPLeases bool               `json:"dhcp_leases"`
	Gravity    PostGravityRequest `json:"gravity"`
}

type PatchConfig struct {
	DNS      map[string]any `json:"dns"`
	DHCP     map[string]any `json:"dhcp"`
	NTP      map[string]any `json:"ntp"`
	Resolver map[string]any `json:"resolver"`
	Database map[string]any `json:"database"`
	Misc     map[string]any `json:"misc"`
	Debug    map[string]any `json:"debug"`
}

type PatchConfigRequest struct {
	Config PatchConfig `json:"config"`
}
