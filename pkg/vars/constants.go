package vars

import "regexp"

var ValidCodePattern = regexp.MustCompile(`^[A-Za-z0-9\-._~!$&'()*+,;=:@]+$`)
