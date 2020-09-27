package util

// Includes will return true if the search string is in the given slice, false otherwise
func Includes(slice []string, search string) bool {
	for i := range slice {
		if slice[i] == search {
			return true
		}
	}
	return false
}
