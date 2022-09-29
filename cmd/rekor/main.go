package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

func main() {
	r, err := sigstore.NewRekor()
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()

	uuids, err := r.SearchByHash(ctx, "40e137b9b9b8204d672642fd1e181c6d5ccb50cfc5cc7fcbb06a8c2c78f44aff")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%v\n", uuids)

	entry, valid, err := r.GetAndVerifyEntry(ctx, uuids[0])
	if err != nil {
		log.Fatalln(err)
	}
	spew.Dump(entry)
	fmt.Printf("IsValid: %v\n", valid)

	rekord, err := r.HashedRekordFromEntry(entry)
	if err != nil {
		log.Fatalln(err)
	}

	spew.Dump(rekord)
	fmt.Printf("key: %s\n", rekord.HashedRekordObj.Signature.PublicKey.Content.String())
}
