package changelog

// NewVersion allocates a new Version struct with all the fields
// initialized except {{.Date}}.
func NewVersion(versionNum string) *Version {
	return &Version{
		Version:     versionNum,
		History:     []*ChangeLine{},
		Subsections: []*Subsection{},
	}
}

// GetVersion fetches the Version struct which matches the versionNum.
// Returns nil if no version was found matching the given versionNum.
func (c *Changelog) GetVersion(versionNum string) *Version {
	for _, v := range c.Versions {
		if v.Version == versionNum {
			return v
		}
	}
	return nil
}

// GetVersion fetches the Version struct which matches the versionNum.
// If no version was found matching the given versionNum, it creates and
// saves it to the Changelog.
func (c *Changelog) GetVersionOrCreate(versionNum string) *Version {
	version := c.GetVersion(versionNum)
	if version == nil {
		version = NewVersion(versionNum)
		c.Versions = append([]*Version{version}, c.Versions...)
	}
	return version
}

// NewSubsection creates a subsection for the given name and initializes its history.
func NewSubsection(subsectionName string) *Subsection {
	return &Subsection{
		Name:    subsectionName,
		History: []*ChangeLine{},
	}
}

// GetSubsection fetches the Subsection struct which matches the versionNum & subsectionName.
// Returns nil if no version was found matching the given versionNum & subsectionName.
func (c *Changelog) GetSubsection(versionNum, subsectionName string) *Subsection {
	version := c.GetVersion(versionNum)
	if version != nil {
		for _, s := range version.Subsections {
			if s.Name == subsectionName {
				return s
			}
		}
	}
	return nil
}

// GetSubsection fetches the Subsection struct which matches the versionNum & subsectionName.
// If no subsection was found matching the given versionNum & subsectionName, it creates it and
// saves it to the Changelog.
func (c *Changelog) GetSubsectionOrCreate(versionNum, subsectionName string) *Subsection {
	version := c.GetVersionOrCreate(versionNum)
	subsection := c.GetSubsection(versionNum, subsectionName)
	if subsection == nil {
		subsection = NewSubsection(subsectionName)
		version.Subsections = append(version.Subsections, subsection)
	}
	return subsection
}

// AddLineToVersion adds a ChangeLine to the given version's direct
// history. This is only to be used when it is inappropriate to add it to a
// subsection, or the version's changes don't warrant subsections.
func (c *Changelog) AddLineToVersion(versionNum string, line *ChangeLine) {
	if line == nil {
		return
	}

	c.addToChangelines(&c.GetVersionOrCreate(versionNum).History, line)
}

// AddLineToSubsection adds a ChangeLine to the given version's
// subsection's history.
//
// For example, this could be used to add a change to v1.4.2's "Major
// Enhancements" subsection.
func (c *Changelog) AddLineToSubsection(versionNum, subsectionName string, line *ChangeLine) {
	if line == nil {
		return
	}

	s := c.GetSubsectionOrCreate(versionNum, subsectionName)
	c.addToChangelines(&s.History, line)
}

// addToChangelines adds a given ChangeLine to the array of ChangeLines.
func (c *Changelog) addToChangelines(lines *[]*ChangeLine, line *ChangeLine) {
	if line == nil || lines == nil {
		return
	}

	*lines = append(*lines, line)
}
