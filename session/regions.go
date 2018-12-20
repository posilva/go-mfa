package session

// Default returns the default regions supported by AWS
// NOTE: in the future this list is subject to change
func DefaultRegions() []string {
	return []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"ap-south-1",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-northeast-1",
		"ca-central-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"sa-east-1",
	}
}
