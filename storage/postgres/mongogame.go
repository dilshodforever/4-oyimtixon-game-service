package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/dilshodforever/4-oyimtixon-game-service/genprotos/game"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GameStorage struct {
	db *mongo.Database
}

func NewGameStorage(db *mongo.Database) *GameStorage {
	return &GameStorage{db: db}
}
func (g *GameStorage) GetLevels(req *game.GetLevelsRequest) (*game.GetLevelsResponse, error) {
	coll := g.db.Collection("levels")
	cursor, err := coll.Find(context.Background(), bson.D{})
	if err != nil {
		log.Printf("Failed to get levels: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var levels []*game.Level
	for cursor.Next(context.Background()) {
		var level game.Level
		if err := cursor.Decode(&level); err != nil {
			log.Printf("Failed to decode level: %v", err)
			return nil, err
		}
		levels = append(levels, &level)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		return nil, err
	}

	return &game.GetLevelsResponse{Levels: levels}, nil
}

func (g *GameStorage) StartLevel(req *game.StartLevelRequest) (*game.StartLevelResponse, error) {
	coll := g.db.Collection("user_levels")
	_, err := coll.InsertOne(context.Background(), bson.D{
		{Key: "user_id", Value: req.Userid},
		{Key: "level_id", Value: req.LevelId},
		{Key: "status", Value: "started"},
		{Key: "user_xp", Value: 0},
	})
	if err != nil {
		log.Printf("Failed to start level: %v", err)
		return nil, err
	}

	return &game.StartLevelResponse{
		Message: "Level started successfully",
	}, nil
}

func (g *GameStorage) CompleteLevel(req *game.CompleteLevelRequest) (*game.CompleteLevelResponse, error) {
	coll := g.db.Collection("user_levels")
	filter := bson.D{
		{Key: "user_id", Value: req.Userid},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: "completed"},
			{Key: "levelid", Value: req.LevelId},
		}},
	}
	_, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Failed to complete level: %v", err)
		return nil, err
	}

	return &game.CompleteLevelResponse{
		Message:          "Level completed successfully",
		XpEarned:         req.Xpearned,
		NewLevelUnlocked: req.LevelId,
	}, nil
}

func (g *GameStorage) GetChallenge(req *game.GetChallengeRequest) (*game.Level, error) {
	coll := g.db.Collection("challenges")
	filter := bson.D{{Key: "id", Value: req.ChallengeId}}
	var challenge game.Challenge
	err := coll.FindOne(context.Background(), filter).Decode(&challenge)
	if err != nil {
		log.Printf("Failed to get challenge: %v", err)
		return nil, err
	}

	levelColl := g.db.Collection("levels")
	var level game.Level
	err = levelColl.FindOne(context.Background(), bson.D{{Key: "challenges.id", Value: req.ChallengeId}}).Decode(&level)
	if err != nil {
		log.Printf("Failed to get level: %v", err)
		return nil, err
	}
	level.Challenges = append(level.Challenges, &challenge)

	return &level, nil
}
func (g *GameStorage) SubmitChallenge(req *game.SubmitChallengeRequest) (*game.SubmitChallengeResponse, error) {
	coll := g.db.Collection("levels")
	filter := bson.D{{Key: "challenges.id", Value: req.ChallengeId}}

	var level game.Level
	err := coll.FindOne(context.Background(), filter).Decode(&level)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("No level found with id: %v", req.ChallengeId)
			return nil, errors.New("no rows in result set")
		}
		log.Printf("Failed to decode level: %v", err)
		return nil, err
	}

	var challenge *game.Challenge
	for _, ch := range level.Challenges {
		if ch.Id == req.ChallengeId {
			challenge = ch
			break
		}
	}
	if challenge == nil {
		log.Printf("Challenge with id: %v not found in level: %v", req.ChallengeId, level.Levelid)
		return nil, errors.New("challenge with id not found in level")
	}

	var submitsresult game.SubmitChallengeResponse
	submitsresult.TotalQuestions = int32(len(challenge.Questions))
	fmt.Println(req.Answers)
	
	for i := 0; i < len(req.Answers); i++ {
		
		for j := 0; j < len(challenge.Questions); j++ {
			fmt.Println(challenge.Questions[j].CorrectOption)
			if req.Answers[i].SelectedOption == challenge.Questions[j].CorrectOption && req.Answers[i].QuestionId == challenge.Questions[j].Id {
				submitsresult.XpEarned += 10
				submitsresult.CorrectAnswers++
			}
		}
	}
	if submitsresult.XpEarned == 0 {
		submitsresult.Feedback = "Keep practicing! You can improve"
		return &submitsresult, nil
	}
	switch submitsresult.CorrectAnswers {
	case int32(len(req.Answers)):
		submitsresult.Feedback = "Excellent! You have a good understanding of quantum superposition."
	case int32(len(req.Answers)) / 2:
		submitsresult.Feedback = "Nice! You're on the right track."
	default:
		submitsresult.Feedback = "Keep practicing! You can improve."
	}

	return &submitsresult, nil
}

func (g *GameStorage) GetLeaderboard(req *game.GetLeaderboardRequest) (*game.LeaderboardResponse, error) {
	coll := g.db.Collection("leaderboard")
	cursor, err := coll.Find(context.Background(), bson.D{})
	if err != nil {
		log.Printf("Failed to get leaderboard: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var leaderboard []*game.LeaderboardEntry
	for cursor.Next(context.Background()) {
		var entry game.LeaderboardEntry
		if err := cursor.Decode(&entry); err != nil {
			log.Printf("Failed to decode leaderboard entry: %v", err)
			return nil, err
		}
		leaderboard = append(leaderboard, &entry)
	}

	userRank := int32(10)

	return &game.LeaderboardResponse{
		Leaderboard: leaderboard,
		UserRank:    userRank,
	}, nil
}

func (g *GameStorage) GetAchievements(req *game.GetAchievementsRequest) (*game.AchievementsResponse, error) {
	coll := g.db.Collection("achievements")
	filter := bson.D{{Key: "user_id", Value: req.Token}}
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to get achievements: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var achievements []*game.Achievement
	for cursor.Next(context.Background()) {
		var achievement game.Achievement
		if err := cursor.Decode(&achievement); err != nil {
			log.Printf("Failed to decode achievement: %v", err)
			return nil, err
		}
		achievements = append(achievements, &achievement)
	}

	return &game.AchievementsResponse{Achievements: achievements}, nil
}

func (g *GameStorage) CheckLevels(req *game.Cheak) (*game.CHeakResult, error) {
	coll := g.db.Collection("levels")
	filter := bson.D{{Key: "id", Value: req.Levelid}}

	var level game.Level
	err := coll.FindOne(context.Background(), filter).Decode(&level)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("No level found with id: %v", req.Levelid)
			return &game.CHeakResult{Result: false}, nil
		}
		log.Printf("Failed to decode level: %v", err)
		return &game.CHeakResult{Result: false}, err
	}

	if level.RequiredXp < req.Userxp {
		xps := level.RequiredXp + level.RequiredXp
		newFilter := bson.D{{Key: "required_xp", Value: xps}}

		var newLevel game.Level
		err := coll.FindOne(context.Background(), newFilter).Decode(&newLevel)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.Printf("No level found with required xp: %v", xps)
				return &game.CHeakResult{Result: false}, nil
			}
			log.Printf("Failed to decode new level: %v", err)
			return &game.CHeakResult{Result: false}, err
		}
		return &game.CHeakResult{Result: true, Levelid: newLevel.Levelid, Xpearned: newLevel.RequiredXp}, nil
	}
	return &game.CHeakResult{Result: false}, nil
}
