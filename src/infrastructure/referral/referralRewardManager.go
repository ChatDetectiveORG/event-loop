package referral

import (
	"context"
	"fmt"
	"log"
	"time"

	e "github.com/ChatDetectiveORG/shared/errors"
	postgresmodels "github.com/ChatDetectiveORG/shared/postgresModels"

	postgresql "github.com/ChatDetectiveORG/event-loop/src/infrastructure/postgresql"
	constants "github.com/ChatDetectiveORG/shared/constants"
)

func StartReferralRewardManagerLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("referral reward manager: stopping")
			return
		case <-ticker.C:
			err := startReferralChecker()
			if e.IsNonNil(err) {
				log.Println(err.JSON())
			}
		}
	}
}

// Selects all money referrals that are eligible for reward and updates users' internal balances
//
// # Fullfils invitor and invited user's internal balances with fixed in referral record reward
//
// Returns: error
func startReferralChecker() *e.ErrorInfo {
	db := postgresql.GetDB()

	var updatedRecordsCount int
	var errors []*e.ErrorInfo

	var validReferrals []postgresmodels.Referral
	err := e.Wrap(db.Model(&validReferrals).
		Where("referral.created_at <= ?", time.Now().Add(-1 * constants.REFERRAL_MONEY_GIVEN_AFTER_SECS*time.Second)).
		Where("referral.fixed_reward_type = ?", postgresmodels.ReferralBonusMoney).
		Relation("Invitor").
		Relation("InvitedUser").
		Select())
	if e.IsNonNil(err) {
		return err.PushStack()
	}

	for _, referral := range validReferrals {
		tx, eRaw := db.Begin()
		if e.IsNonNil(eRaw) {
			errors = append(errors, e.FromError(eRaw, "failed to begin transaction"))
			continue
		}
		defer tx.Rollback()

		invitor := referral.Invitor
		invited := referral.InvitedUser

		if invitor == nil {
			eRaw := tx.Model(&invitor).Where("id = ?", referral.InvitorID).Select()
			if e.IsNonNil(eRaw) {
				errors = append(errors, e.FromError(eRaw, "failed to get invitor via referral record. Maybe user deleted data?"))
			}
		}

		if invited == nil {
			eRaw := tx.Model(&invited).Where("id = ?", referral.InvitedUserID).Select()
			if e.IsNonNil(eRaw) {
				errors = append(errors, e.FromError(eRaw, "failed to get invited user via referral record. Maybe user deleted data?"))
			}
		}

		if invitor != nil {
			invitor.InternalBalance += referral.FixedMoneyReward
			invitor.InternalBalanceUpdatedAt = time.Now()
			_, eRaw := tx.Model(invitor).WherePK().Column("internal_balance", "internal_balance_updated_at").Update()
			if e.IsNonNil(eRaw) {
				errors = append(errors, e.FromError(eRaw, "failed to update invitor internal balance"))
			}
		}

		if invited != nil {
			invited.InternalBalance += referral.FixedMoneyReward
			invited.InternalBalanceUpdatedAt = time.Now()
			_, eRaw := tx.Model(invited).WherePK().Column("internal_balance", "internal_balance_updated_at").Update()
			if e.IsNonNil(eRaw) {
				errors = append(errors, e.FromError(eRaw, "failed to update invited user internal balance"))
			}
		}

		_, eRaw = tx.Model(&referral).WherePK().Delete()
		if e.IsNonNil(eRaw) {
			errors = append(errors, e.FromError(eRaw, "failed to delete referral record"))
		}

		if eRaw = tx.Commit(); e.IsNonNil(eRaw) {
			errors = append(errors, e.FromError(eRaw, "failed to commit transaction"))
			continue
		}

		updatedRecordsCount++
	}

	log.Println("Referral reward manager: updated", updatedRecordsCount)

	if len(errors) > 0 {
		err = e.NewError("failed_to_update_some_referral_records", "Failed to update some referral records")
		for i, err_i := range errors {
			err.Data[fmt.Sprintf("error_%d", i)] = err_i.Error()
		}

		return err.PushStack()
	}

	return e.Nil()
}
