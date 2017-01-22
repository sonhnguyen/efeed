package dotastats

import "strings"

func GetTeamMatches(teamName, limit string, mongodb Mongodb) ([]Match, error) {
	result, err := mongodb.GetTeamMatches(teamName, limit)
	if err != nil {
		return []Match{}, err
	}

	return result, nil
}

func GetTeamF10kMatches(teamName, limit string, mongodb Mongodb) ([]Match, error) {
	result, err := mongodb.GetTeamF10kMatches(teamName, limit)
	if err != nil {
		return []Match{}, err
	}

	return result, nil
}

func GetF10kResult(teamName, limit string, mongodb Mongodb) (F10kResult, error) {
	data, err := mongodb.GetTeamF10kMatches(teamName, limit)
	if err != nil {
		return F10kResult{}, err
	}
	if len(data) == 0 {
		return F10kResult{}, nil
	}

	var result F10kResult
	var enemyList []string
	var kill, death, ratio, totalKill, totalDeath, win, avgKill, avgDeath, winrate, avgOdds, ratioKill float64
	for _, match := range data {
		var enemy string
		if match.ScoreA == 0 {
			match.ScoreA = 1
		}
		if match.ScoreB == 0 {
			match.ScoreB = 1
		}
		if teamName == strings.ToLower(match.TeamA) {
			kill = float64(match.ScoreA)
			death = float64(match.ScoreB)
			ratio = float64(match.RatioA)
			enemy = strings.ToLower(match.TeamB)
		} else {
			kill = float64(match.ScoreB)
			death = float64(match.ScoreA)
			ratio = float64(match.RatioB)
			enemy = strings.ToLower(match.TeamA)
		}
		totalKill += kill
		totalDeath += death
		if teamName == strings.ToLower(match.Winner) {
			win++
		}
		avgOdds += ratio
		ratioKill += kill / death
		enemyList = append(enemyList, enemy)
	}
	avgKill = totalKill / float64(len(data))
	avgDeath = totalDeath / float64(len(data))
	winrate = win / float64(len(data))
	avgOdds = avgOdds / float64(len(data))
	ratioKill = ratioKill / float64(len(data))
	result = F10kResult{AverageDeath: avgDeath, AverageKill: avgKill, Name: teamName, RatioKill: ratioKill, TotalKill: totalKill, TotalDeath: totalDeath, Winrate: winrate, AverageOdds: avgOdds, Enemy: enemyList}
	return result, nil
}
