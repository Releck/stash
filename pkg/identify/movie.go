package identify

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/plugin/common/log"
	"github.com/stashapp/stash/pkg/utils"
)

func getMovie(ctx context.Context, r models.Repository, m *models.ScrapedMovie, createMissing bool, studioID *int64) (*models.Movie, error) {
	var movie *models.Movie
	var err error
	if createMissing && m.StoredID == nil {
		movie, err = r.Movie().Create(*scrapedToMovieInput(ctx, r, m, studioID))
		if err != nil {
			return nil, fmt.Errorf("error creating movie", err)
		}
		err = createCoverImages(ctx, r, m, movie)
	} else {
		movie, err = r.Movie().UpdateFull(*scrapedToMovieInput(ctx, r, m, studioID))
		if err != nil {
			return nil, fmt.Errorf("error updating movie", err)
		}
		return movie, nil
	}

	return movie, err
}

func createCoverImages(ctx context.Context, r models.Repository, sm *models.ScrapedMovie, m *models.Movie) error {
	var frontImageData []byte
	var backimageData []byte
	var err error
	if sm.FrontImage == nil && sm.BackImage != nil {
		sm.FrontImage = &models.DefaultMovieImage
	}

	if sm.FrontImage != nil {
		frontImageData, err = utils.ProcessImageInput(ctx, *sm.FrontImage)
		if err != nil {
			fmt.Errorf("error processing front movie image: %w", err)
		}
	}

	if sm.BackImage != nil {
		backimageData, err = utils.ProcessImageInput(ctx, *sm.BackImage)
		if err != nil {
			fmt.Errorf("error processing back movie image: %w", err)
		}
	}

	if len(frontImageData) > 0 {
		if err := r.Movie().UpdateImages(m.ID, frontImageData, backimageData); err != nil {
			return fmt.Errorf("error updating images: %w", err)
		}
	}
	return nil
}

func scrapedToMovieInput(ctx context.Context, r models.Repository, movie *models.ScrapedMovie, studioID *int64) *models.Movie {
	ret := models.NewMovie(*movie.Name)

	if movie.StoredID != nil {
		mID, err := strconv.Atoi(*movie.StoredID)
		if err != nil {
			log.Warn("error parsing movie studioID: %w", err)
			return nil
		}
		ret.ID = mID
	}

	if movie.Aliases != nil {
		ret.Aliases = sql.NullString{String: *movie.Aliases, Valid: true}
	}
	if movie.Date != nil {
		ret.Date = models.SQLiteDate{String: *movie.Date, Valid: true}
	}
	if movie.Duration != nil {
		duration, err := strconv.ParseInt(*movie.Duration, 10, 64)
		if err == nil {
			ret.Duration = sql.NullInt64{Int64: int64(duration), Valid: true}
		}
	}
	if movie.Director != nil {
		ret.Director = sql.NullString{String: *movie.Director, Valid: true}
	}
	if movie.URL != nil {
		ret.URL = sql.NullString{String: *movie.URL, Valid: true}
	}
	if movie.Synopsis != nil {
		ret.Synopsis = sql.NullString{String: *movie.Synopsis, Valid: true}
	}

	var sID int64
	var err error
	if studioID != nil {
		sID = *studioID
	} else {
		studioID, err = createMissingStudio("", r, movie.Studio)
		if err != nil {
			log.Warn("error creating studio", err)
		}
	}
	if err == nil {
		ret.StudioID = sql.NullInt64{Int64: sID, Valid: true}
	}

	return ret
}
