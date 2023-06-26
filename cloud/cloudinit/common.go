package cloudinit

func IsValidType(cloudInitType string) bool {
	switch cloudInitType {
	case "user", "meta", "network":
		return true
	default:
		return false
	}
}
