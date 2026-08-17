package api

func CanReviewSeller(role string) bool {
	return role == "operator" || role == "admin"
}
