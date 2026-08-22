package tinygo

func EnsureInstalled(opts ...Option) (string, error) {
	c := newConfig(opts...)

	// Check if the correct version is already reachable via PATH.
	if p, err := getPath(c); err == nil {
		ver, verErr := installedVersion(opts...)
		if verErr == nil && ver == c.version {
			c.logger("TinyGo " + ver + " already installed at " + p)
			return p, nil
		}
		// Version mismatch — remove the old install before installing the new one.
		c.logger("TinyGo version mismatch: have " + versionOrUnknown(ver, verErr) + ", want " + c.version + ". Removing old install...")
		if err := removeExisting(c, p); err != nil {
			return "", err
		}
	}

	if err := Install(opts...); err != nil {
		return "", err
	}

	return getPath(c)
}

func versionOrUnknown(ver string, err error) string {
	if err != nil {
		return "unknown"
	}
	return ver
}
