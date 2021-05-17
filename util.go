package main

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func levenshtein(s1, s2 string) int {
	a, b := []rune(s1), []rune(s2)

	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)

		for j := range dp[i] {
			dp[i][j] = 0
		}
	}

	dp[0][0] = 0
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			if a[i] == b[j] {
				dp[i+1][j+1] = max(dp[i+1][j+1], dp[i][j]+1)
			} else {
				dp[i+1][j+1] = max(dp[i+1][j+1], dp[i][j])
			}

			dp[i][j+1] = max(dp[i][j+1], dp[i][j])
			dp[i+1][j] = max(dp[i+1][j], dp[i][j])
		}
	}

	return dp[len(dp)-1][len(dp[0])-1]
}
