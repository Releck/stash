package identify

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/stashapp/stash/pkg/hash/md5"
	"github.com/stashapp/stash/pkg/models"
)

func getPerformerID(endpoint string, r models.Repository, p *models.ScrapedPerformer, createMissing bool) (*int, error) {
	if p.StoredID != nil {
		// existing performer, just add it
		performerID, err := strconv.Atoi(*p.StoredID)
		if err != nil {
			return nil, fmt.Errorf("error converting performer ID %s: %w", *p.StoredID, err)
		}

		return &performerID, nil
	} else if createMissing && p.Name != nil { // name is mandatory
		return createMissingPerformer(endpoint, r, p)
	}

	return nil, nil
}

func createPerformerImage(ctx context.Context, r models.Repository, p *models.ScrapedPerformer, performerID int) (error) {
	if p.Image != nil {
		performerImage, err := utils.ReadImageFromURL(ctx, *p.Image)
		if err != nil {
			return fmt.Errorf("error reading performer image: %w", err)
		}
		if len(performerImage) > 0 {
			err = r.Performer().UpdateImage(performerID, performerImage)
			if err != nil {
				return fmt.Errorf("error updating performer image: %w", err)
			}
		}
	}
	return nil
}

func createPerformerTags(ctx context.Context, r models.Repository, p *models.ScrapedPerformer, performerID int) (error) {
	var newTags []int
	var tID int
	var err error
	var newTag *models.Tag
	for _, tag := range p.Tags {
		err = nil
		if tag.StoredID != nil {
			tID, err = strconv.Atoi(*tag.StoredID)
			if err == nil {
				newTag = &models.Tag{
					ID: tID,
					Name: tag.Name,
				}
			}
		} else {
			newTag, err = r.Tag().Create(*models.NewTag(tag.Name))
			if err != nil {
				return fmt.Errorf("error creating tag %v: %v", tag.Name, err)
			}
		}
		if err == nil {
			newTags = append(newTags, newTag.ID)
		}
	}
	if len(newTags) > 0 {
		err = r.Performer().UpdateTags(performerID, newTags)
		if err != nil {
			return fmt.Errorf("error updating performer tags: %v", err)
		}
	}
	return nil
}

func createMissingPerformer(endpoint string, r models.Repository, p *models.ScrapedPerformer) (*int, error) {
	created, err := r.Performer().Create(scrapedToPerformerInput(p))
	if err != nil {
		return nil, fmt.Errorf("error creating performer: %w", err)
	}

	if endpoint != "" && p.RemoteSiteID != nil {
		if err := r.Performer().UpdateStashIDs(created.ID, []models.StashID{
			{
				Endpoint: endpoint,
				StashID:  *p.RemoteSiteID,
			},
		}); err != nil {
			return nil, fmt.Errorf("error setting performer stash id: %w", err)
		}
	}

	return &created.ID, nil
}

func scrapedToPerformerInput(performer *models.ScrapedPerformer) models.Performer {
	currentTime := time.Now()
	ret := models.Performer{
		Name:      sql.NullString{String: *performer.Name, Valid: true},
		Checksum:  md5.FromString(*performer.Name),
		CreatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
		UpdatedAt: models.SQLiteTimestamp{Timestamp: currentTime},
		Favorite:  sql.NullBool{Bool: false, Valid: true},
	}
	if performer.Birthdate != nil {
		ret.Birthdate = models.SQLiteDate{String: *performer.Birthdate, Valid: true}
	}
	if performer.DeathDate != nil {
		ret.DeathDate = models.SQLiteDate{String: *performer.DeathDate, Valid: true}
	}
	if performer.Gender != nil {
		ret.Gender = sql.NullString{String: *performer.Gender, Valid: true}
	}
	if performer.Ethnicity != nil {
		ret.Ethnicity = sql.NullString{String: *performer.Ethnicity, Valid: true}
	}
	if performer.Country != nil {
		ret.Country = sql.NullString{String: *performer.Country, Valid: true}
	}
	if performer.EyeColor != nil {
		ret.EyeColor = sql.NullString{String: *performer.EyeColor, Valid: true}
	}
	if performer.HairColor != nil {
		ret.HairColor = sql.NullString{String: *performer.HairColor, Valid: true}
	}
	if performer.Height != nil {
		ret.Height = sql.NullString{String: *performer.Height, Valid: true}
	}
	if performer.Measurements != nil {
		ret.Measurements = sql.NullString{String: *performer.Measurements, Valid: true}
	}
	if performer.FakeTits != nil {
		ret.FakeTits = sql.NullString{String: *performer.FakeTits, Valid: true}
	}
	if performer.CareerLength != nil {
		ret.CareerLength = sql.NullString{String: *performer.CareerLength, Valid: true}
	}
	if performer.Tattoos != nil {
		ret.Tattoos = sql.NullString{String: *performer.Tattoos, Valid: true}
	}
	if performer.Piercings != nil {
		ret.Piercings = sql.NullString{String: *performer.Piercings, Valid: true}
	}
	if performer.Aliases != nil {
		ret.Aliases = sql.NullString{String: *performer.Aliases, Valid: true}
	}
	if performer.Twitter != nil {
		ret.Twitter = sql.NullString{String: *performer.Twitter, Valid: true}
	}
	if performer.Instagram != nil {
		ret.Instagram = sql.NullString{String: *performer.Instagram, Valid: true}
	}

	return ret
}