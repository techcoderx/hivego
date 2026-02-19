package hivego

import "time"

func getTestVoteOp() HiveOperation {
	return voteOperation{
		Voter:    "xeroc",
		Author:   "xeroc",
		Permlink: "piston",
		Weight:   10000,
		opText:   "vote",
	}
}

func getTestCustomJsonOp() HiveOperation {
	return CustomJsonOperation{
		RequiredAuths:        []string{},
		RequiredPostingAuths: []string{"xeroc"},
		Id:                   "test-id",
		Json:                 "{\"testk\":\"testv\"}",
		opText:               "custom_json",
	}
}

func getTestAccCreateOp() HiveOperation {
	return AccountCreateOperation{
		Fee:            "0.000 HIVE",
		Creator:        "milo-hpr",
		NewAccountName: "sagar",
		Owner: Auths{
			WeightThreshold: 1,
			KeyAuths:        [][2]interface{}{{"STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow", 1}},
		},
		Active: Auths{
			WeightThreshold: 1,
			KeyAuths:        [][2]interface{}{{"STM6e1heeScT5oj8AsYKdRGfYcqiqbiZkpWY8qL3uuHZY4mLPjiYb", 1}},
		},
		Posting: Auths{
			WeightThreshold: 1,
			KeyAuths:        [][2]interface{}{{"STM8VPjfDcioxjkc5dRK8oi4jiyKagEZmvL9pCmmm4M9utkFh2SbK", 1}},
		},
		MemoKey:      "STM67xAW8SFTki89r2TfFBZcKPRyRjL2D9kWd5wf1FFKdQULeqiPu",
		JsonMetadata: "",
	}
}

func getTestAccountUpdateOp() HiveOperation {
	return AccountUpdateOperation{
		Account:      "sniperduel17",
		Owner:        nil,
		Active:       nil,
		Posting:      nil,
		MemoKey:      "STM6n4WcwyiC63udKYR8jDFuzG9T48dhy2Qb5sVmQ9MyNuKM7xE29",
		JsonMetadata: "{\"foo\":\"bar\"}",
		opText:       "account_update",
	}
}

func getTestTransferOp() HiveOperation {
	// legacy: 1d59a3a21cc704686a942c513658463fac61561a
	// hf26: 2e40400aa62dc4b967e2c5b703a86f1e23ab4907
	return TransferOperation{
		To:     "vsc.gateway",
		From:   "tibfox.vsc",
		Memo:   "to=tibfox",
		Amount: "1.000 HIVE",
	}
}

func getTwoTestOps() []HiveOperation {
	return []HiveOperation{getTestVoteOp(), getTestCustomJsonOp()}
}

func getTestTx(ops []HiveOperation) HiveTransaction {
	exp, _ := time.Parse("2006-01-02T15:04:05", "2016-08-08T12:24:17")
	expStr := exp.Format("2006-01-02T15:04:05")

	return HiveTransaction{
		RefBlockNum:    36029,
		RefBlockPrefix: 1164960351,
		Expiration:     expStr,
		Operations:     ops,
	}
}

func getTestVoteTx() HiveTransaction {
	return getTestTx([]HiveOperation{getTestVoteOp()})
}

func getTestClaimAcc() HiveOperation {
	return ClaimAccountOperation{
		Fee:     "0.000 HIVE",
		Creator: "techcoderx",
	}
}
