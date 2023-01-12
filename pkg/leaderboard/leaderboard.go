package leaderboard

import (
	"badger-api/pkg/drill"
	"badger-api/pkg/server"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const LeaderboardOverallScoreCollection = "leaderboard_scores"

type LeaderboardPlayerDoc struct {
	UserId string `bson:"_id"`
	Name   string `bson:"name"`
	Score  uint32 `bson:"score"`
}

type LeaderboardPlayer struct {
	UserId     string
	Name       string
	TotalScore uint32
	Breakdown  PlayerScore
}

type PlayerScore struct {
	BattingScore  uint32
	CatchingScore uint32
	BowlingScore  uint32

	TotalBattingSubmissions  uint32
	TotalCatchingSubmissions uint32
	TotalBowlingSubmissions  uint32
}

func GetPlayerScore(s *server.ServerContext, userId string) PlayerScore {
	return PlayerScore{
		BattingScore:  drill.ComputeBattingScoreForUser(s, userId),
		CatchingScore: 0,
		BowlingScore:  0,

		TotalBattingSubmissions:  drill.CountBattingSubmissionsByUser(s, userId),
		TotalCatchingSubmissions: 0,
		TotalBowlingSubmissions:  0,
	}
}

func GetTopPlayers(s *server.ServerContext, count int) []LeaderboardPlayer {
	col := s.GetCollection(LeaderboardOverallScoreCollection)

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "score", Value: -1}})

	cursor, err := col.Find(s.GetMongoContext(), filter, opts)

	var results []LeaderboardPlayerDoc
	if err = cursor.All(s.GetMongoContext(), &results); err != nil {
		panic(err)
	}

	var leaderboard []LeaderboardPlayer
	for i, result := range results {
		if i == count {
			break
		}

		leaderboard = append(leaderboard, LeaderboardPlayer{
			UserId:     result.UserId,
			Name:       result.Name,
			TotalScore: result.Score,
			Breakdown:  GetPlayerScore(s, result.UserId),
		})
	}

	return leaderboard
}

func UpdatePlayerLeaderboardScore(s *server.ServerContext, userId string) {
	score := GetPlayerScore(s, userId)
	overallScore := (score.BattingScore + score.BowlingScore + score.CatchingScore) / 3

	col := s.GetCollection(LeaderboardOverallScoreCollection)

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "score", Value: uint32(overallScore)}}}}
	opts := options.Update().SetUpsert(true)

	col.UpdateByID(s.GetMongoContext(), userId, update, opts)
}

func UpdatePlayerLeaderboardName(s *server.ServerContext, userId, name string) {
	col := s.GetCollection(LeaderboardOverallScoreCollection)

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: name}}}}
	opts := options.Update().SetUpsert(true)

	col.UpdateByID(s.GetMongoContext(), userId, update, opts)
}