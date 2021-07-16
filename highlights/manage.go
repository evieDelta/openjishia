package highlights

func setguildenable(guildID string, state bool) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[guildID] == nil {
		g := globalruntime.Guildsettings[guildID]
		g = &guildsettings{}
		globalruntime.Guildsettings[guildID] = g
	}

	globalruntime.Guildsettings[guildID].Enable = state
}

func isguildenabled(guildID string) bool {
	globallock.RLock()
	defer globallock.RUnlock()

	if globalruntime.Guildsettings[guildID] == nil {
		return false
	}

	return globalruntime.Guildsettings[guildID].Enable
}

func useraddhl(userID, guildID string, term string) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[guildID] == nil {
		globalruntime.Guildsettings[guildID] = &guildsettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID] == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID] = &membersettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords = make(map[string]*highlight)
	}

	// fmt.Println(globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords[term])

	hld := globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords[term]
	hld = &highlight{}
	globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords[term] = hld

	// fmt.Println(*globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords[term])
}

func userremhl(userID, guildID string, term string) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[guildID] == nil {
		globalruntime.Guildsettings[guildID] = &guildsettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID] == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID] = &membersettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords = make(map[string]*highlight)
	}

	delete(globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords, term)
}

func userlshl(userID, guildID string) []string {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[guildID] == nil {
		globalruntime.Guildsettings[guildID] = &guildsettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID] == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID] = &membersettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords = make(map[string]*highlight)
	}

	list := make([]string, 0, len(globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords))
	for x := range globalruntime.Guildsettings[guildID].MemberSettings[userID].Highlightwords {
		list = append(list, x)
	}

	return list
}

func useraddhlblock(userID, guildID string, term string, typ int) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[guildID] == nil {
		globalruntime.Guildsettings[guildID] = &guildsettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID] == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID] = &membersettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks = make(map[string]blockerstuff)
	}

	globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks[term] = blockerstuff{State: true, Thing: typ}
}

func userremhlblock(userID, guildID string, term string) {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[guildID] == nil {
		globalruntime.Guildsettings[guildID] = &guildsettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID] == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID] = &membersettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks = make(map[string]blockerstuff)
	}

	delete(globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks, term)
}

func userlshlblock(userID, guildID string) map[string]blockerstuff {
	globallock.Lock()
	defer globallock.Unlock()

	if globalruntime.Guildsettings[guildID] == nil {
		globalruntime.Guildsettings[guildID] = &guildsettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID] == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID] = &membersettings{}
	}
	if globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks == nil {
		globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks = make(map[string]blockerstuff)
	}

	return globalruntime.Guildsettings[guildID].MemberSettings[userID].Blocks
}
