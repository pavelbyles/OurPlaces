package backend

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// Listings contains an array of all listing
type Listings struct {
	Listings []Listing `json:"listings"`
	Next     string    `json:"next"`
}

// ResultResponse is a generic response type
type ResultResponse struct {
	IsSuccessful bool `json:"isSuccessful"`
}

// Listing is an listing item
type Listing struct {
	ID                 int64              `json:"id" datastore:"-"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	PropertyType       string             `json:"propertyType"`
	RoomType           string             `json:"roomType"`
	NumGuests          int32              `json:"numGuests"`
	NumBedrooms        int32              `json:"numBedrooms"`
	NumBeds            int32              `json:"numBeds"`
	IsActive           bool               `json:"isActive"`
	NumStars           int8               `json:"numStars"`
	BusinessAttr       BusinessAttributes `json:"businessAttributes"`
	SelfCheckInMethods SelfCheckinMethod  `json:"selfCheckInMethods"`
}

// BusinessAttributes has all attributes for business-type listings
type BusinessAttributes struct {
	HasEssentials      bool `json:"hasEssentials"`
	HasShampoo         bool `json:"hasShampoo"`
	HasWifi            bool `json:"hasWifi"`
	HasHangers         bool `json:"hasHangers"`
	HasIron            bool `json:"hasIron"`
	HasHairDryer       bool `json:"hasHairDryer"`
	HasLaptopWorkspace bool `json:"hasLaptopWorkspace"`
	HasSmokeDetector   bool `json:"hasSmokeDetector"`
	HasCoDetector      bool `json:"hasCoDetector"`
	IsNoSmoking        bool `json:"isNoSmoking"`
	IsNoPets           bool `json:"isNoPets"`
}

// SelfCheckinMethod has all attributes for self-checkin
type SelfCheckinMethod struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Key is the key for listings
func (listing *Listing) Key(c context.Context) *datastore.Key {
	// return datastore.NewKey(c, "Listing", listing.Name, 0, listingKey(c))
	return datastore.NewIncompleteKey(c, "Listing", nil)
}

// AddListing saves a new listing
func AddListing(c context.Context, listing *Listing) (*datastore.Key, error) {
	key, err := datastore.Put(c, listing.Key(c), listing)
	if nil != err {
		return nil, err
	}
	return key, updateCounter(c, "Listing")
}

// GetAllListings retrieves all listings based on the window specified
func GetAllListings(c context.Context, pageSize int, next string) ([]Listing, string, error) {
	q := datastore.NewQuery("Listing").Limit(pageSize)
	listings := make([]Listing, 0)

	if "" != next {
		cursor, err := datastore.DecodeCursor(next)
		if err == nil {
			q = q.Start(cursor)
		}
	}

	isDone := false
	it := q.Run(c)
	for i := 0; i < pageSize; i++ {
		var listing Listing
		key, err := it.Next(&listing)
		if err == datastore.Done {
			log.Infof(c, "No more results")
			isDone = true
			break
		}
		if err != nil {
			log.Errorf(c, "Fetching next listing: %v", err)
			break
		}
		listing.ID = key.IntID()
		listings = append(listings, listing)
	}

	if isDone {
		return listings, "", nil
	}

	cursor, err := it.Cursor()
	if err == nil {
		return listings, cursor.String(), nil
	}
	return listings, "", err
}

// DeleteListing deletes a listing from the datastore
func DeleteListing(c context.Context, key *datastore.Key) error {
	err := datastore.Delete(c, key)
	return err
}

// GetListingByKey returns single listing based on string key
func GetListingByKey(c context.Context, id int64) (*Listing, error) {
	listingKey := datastore.NewKey(c, "Listing", "", id, nil)

	var listing Listing
	err := datastore.Get(c, listingKey, &listing)
	if nil != err {
		log.Errorf(c, "Could not retrieve listing with specified ID (%v): %v", id, err)
		return nil, err
	}
	listing.ID = id

	return &listing, nil
}
