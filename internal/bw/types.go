package bw

import "time"

type Item struct {
	Object         string          `json:"object"`
	ID             string          `json:"id"`
	OrganizationID *string         `json:"organizationId"`
	FolderID       *string         `json:"folderId"`
	Type           int             `json:"type"`
	Name           string          `json:"name"`
	Notes          *string         `json:"notes"`
	Favorite       bool            `json:"favorite"`
	Reprompt       int             `json:"reprompt"`
	CollectionIDs  []string        `json:"collectionIds"`
	RevisionDate   time.Time       `json:"revisionDate"`
	CreationDate   time.Time       `json:"creationDate"`
	DeletedDate    *time.Time      `json:"deletedDate"`
	Fields         []CustomField   `json:"fields"`
	Login          *Login          `json:"login"`
	Card           *Card           `json:"card"`
	Identity       *Identity       `json:"identity"`
	SecureNote     *SecureNote     `json:"secureNote"`
}

type Login struct {
	URIs                 []URI      `json:"uris"`
	Username             *string    `json:"username"`
	Password             *string    `json:"password"`
	PasswordRevisionDate *time.Time `json:"passwordRevisionDate"`
	TOTP                 *string    `json:"totp"`
}

type URI struct {
	URI   string `json:"uri"`
	Match *int   `json:"match"`
}

type Card struct {
	CardholderName *string `json:"cardholderName"`
	Brand          *string `json:"brand"`
	Number         *string `json:"number"`
	ExpMonth       *string `json:"expMonth"`
	ExpYear        *string `json:"expYear"`
	Code           *string `json:"code"`
}

type Identity struct {
	FirstName      *string `json:"firstName"`
	MiddleName     *string `json:"middleName"`
	LastName       *string `json:"lastName"`
	Username       *string `json:"username"`
	Company        *string `json:"company"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	SSN            *string `json:"ssn"`
	PassportNumber *string `json:"passportNumber"`
	LicenseNumber  *string `json:"licenseNumber"`
	Address1       *string `json:"address1"`
	Address2       *string `json:"address2"`
	Address3       *string `json:"address3"`
	City           *string `json:"city"`
	State          *string `json:"state"`
	PostalCode     *string `json:"postalCode"`
	Country        *string `json:"country"`
}

type SecureNote struct {
	Type int `json:"type"`
}

type CustomField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

type Folder struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

type Status struct {
	ServerURL   string  `json:"serverUrl"`
	LastSync    *string `json:"lastSync"`
	UserEmail   *string `json:"userEmail"`
	UserID      *string `json:"userId"`
	Status      string  `json:"status"` // "unauthenticated", "locked", "unlocked"
}

const (
	ItemTypeLogin      = 1
	ItemTypeSecureNote = 2
	ItemTypeCard       = 3
	ItemTypeIdentity   = 4
)
