package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/repositories"
	pb "github.com/kisanlink/protobuf/pb-aaa"
)

// KisansathiServiceInterface exposes the public contract
type KisansathiServiceInterface interface {
	ListKisansathis(fpoRegNo string) ([]*pb.User, error)
}

type KisansathiService struct {
	farmerRepo repositories.FarmerRepositoryInterface
}

func NewKisansathiService(repo repositories.FarmerRepositoryInterface) *KisansathiService {
	return &KisansathiService{farmerRepo: repo}
}

func (s *KisansathiService) ListKisansathis(fpoRegNo string) ([]*pb.User, error) {
	// 1. Fetch farmers filtered by optional FPO reg‑no
	farmers, err := s.farmerRepo.FetchFarmers(
		"", // userId
		"", // farmerId
		"", // kisansathiUserId
		fpoRegNo,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch farmers: %w", err)
	}
	if len(farmers) == 0 {
		return nil, nil // caller treats as "empty list"
	}

	// 2. Collect distinct KisansathiUserIds
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for _, f := range farmers {
		if f.KisansathiUserId != nil && *f.KisansathiUserId != "" {
			if _, ok := seen[*f.KisansathiUserId]; !ok {
				seen[*f.KisansathiUserId] = struct{}{}
				ids = append(ids, *f.KisansathiUserId)
			}
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}

	// 3. Call AAA for each ID; ignore individual failures but log them
	out := make([]*pb.User, 0, len(ids))
	for _, id := range ids {
		u, err := GetUserByIdClient(context.Background(), id)
		if err != nil {
			log.Printf("kisansathi_service: failed AAA lookup for %s: %v", id, err)
			continue
		}
		if u != nil && u.Data != nil {
			out = append(out, u.Data)
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}
