package demo

import (
	"fmt"
	"time"

	"github.com/owenrumney/bwql/internal/bw"
)

type Client struct {
	nextFolderID int
}

func NewClient() *Client {
	return &Client{nextFolderID: 100}
}

func (c *Client) CreateFolder(name string) (*bw.Folder, error) {
	c.nextFolderID++
	return &bw.Folder{
		ID:   fmt.Sprintf("folder-%d", c.nextFolderID),
		Name: name,
	}, nil
}

func (c *Client) EditFolder(id, name string) (*bw.Folder, error) {
	return &bw.Folder{ID: id, Name: name}, nil
}

func (c *Client) DeleteFolder(_ string) error {
	return nil
}

func (c *Client) EditItem(item *bw.Item) (*bw.Item, error) {
	item.RevisionDate = time.Now()
	return item, nil
}

func (c *Client) DeleteItem(_ string) error {
	return nil
}

func Folders() []bw.Folder {
	return []bw.Folder{
		{ID: "folder-1", Name: "Work"},
		{ID: "folder-2", Name: "Personal"},
		{ID: "folder-3", Name: "Finance"},
		{ID: "folder-4", Name: "Development"},
		{ID: "folder-5", Name: "Social"},
	}
}

func Items() []bw.Item {
	now := time.Now()

	d := func(daysAgo int) time.Time { return now.AddDate(0, 0, -daysAgo) }
	s := func(v string) *string { return &v }
	t := func(v time.Time) *time.Time { return &v }
	f := func(id string) *string { return &id }

	return []bw.Item{
		// Work folder - logins
		login("1", "GitHub Enterprise", f("folder-1"), s("charlie.brown@acme.com"), s("Kj8$mNp2!xQw9#Lz"), "https://github.acme.com", s("JBSWY3DPEHPK3PXP"), t(d(15)), d(15), d(400)),
		login("2", "AWS Console", f("folder-1"), s("charlie.brown"), s("r5&Yt!mK8@pW2nXq"), "https://aws.amazon.com", s("NBSWY3DPEHPK3PXQ"), t(d(45)), d(45), d(800)),
		login("3", "Jira", f("folder-1"), s("charlie.brown@acme.com"), s("Hm4$kLp8!wQx2#Nz"), "https://acme.atlassian.net", nil, t(d(500)), d(120), d(900)),
		login("4", "Slack", f("folder-1"), s("charlie.brown@acme.com"), s("Wq9!nMk3$xPz8#Lt"), "https://acme.slack.com", s("KBSWY3DPEHPK3PXR"), t(d(30)), d(30), d(600)),
		login("5", "Confluence", f("folder-1"), s("charlie.brown@acme.com"), s("Yt7&mKp5!wNx2$Qz"), "https://acme.atlassian.net/wiki", nil, t(d(450)), d(90), d(900)),
		login("6", "GitLab", f("folder-1"), s("charlie"), s("Xp3!kLm8$wQz9#Nt"), "https://gitlab.acme.com", nil, t(d(380)), d(60), d(700)),
		login("7", "Office 365", f("folder-1"), s("charlie.brown@acme.com"), s("Nz6$mWp4!xKt8#Lq"), "https://login.microsoftonline.com", s("MBSWY3DPEHPK3PXS"), t(d(20)), d(20), d(500)),

		// Personal folder - logins
		login("8", "Gmail", f("folder-2"), s("charlie.brown@gmail.com"), s("Km8!nYp3$wQx5#Lt"), "https://mail.google.com", s("OBSPY3DPEHPK3PXT"), t(d(10)), d(10), d(1200)),
		login("9", "Netflix", f("folder-2"), s("charlie.brown@gmail.com"), s("Qw2$mKp7!xNz9#Yt"), "https://netflix.com", nil, t(d(600)), d(200), d(900)),
		login("10", "Spotify", f("folder-2"), s("cbrown"), s("Lt5!kMp8$wXz3#Nq"), "https://accounts.spotify.com", nil, t(d(730)), d(300), d(1100)),
		login("11", "Amazon", f("folder-2"), s("charlie.brown@gmail.com"), s("Wz8$nKp4!xQt2#Lm"), "https://amazon.co.uk", s("PBSWY3DPEHPK3PXU"), t(d(25)), d(25), d(1000)),
		login("12", "Reddit", f("folder-2"), s("charlie_b"), s("password123"), "https://reddit.com", nil, t(d(900)), d(400), d(1400)),
		login("13", "Twitter/X", f("folder-2"), s("cbrown"), s("Xq7!mLp2$wNt8#Kz"), "https://x.com", nil, t(d(550)), d(180), d(800)),
		login("14", "Apple ID", f("folder-2"), s("charlie.brown@icloud.com"), s("Yt4$kWp9!xQz3#Nm"), "https://appleid.apple.com", s("QBSWY3DPEHPK3PXV"), t(d(60)), d(60), d(1500)),

		// Finance folder - logins
		login("15", "Barclays Online Banking", f("folder-3"), s("cbrown"), s("Nq6!mKp3$wXt8#Lz"), "https://online.barclays.co.uk", s("RBSWY3DPEHPK3PXW"), t(d(90)), d(90), d(1800)),
		login("16", "Monzo", f("folder-3"), s("charlie.brown@gmail.com"), s("Km9$nWp5!xQz2#Yt"), "https://monzo.com", s("SBSWY3DPEHPK3PXX"), t(d(30)), d(30), d(600)),
		login("17", "Coinbase", f("folder-3"), s("charlie.brown@gmail.com"), s("Wt3!kMp7$xNz9#Lq"), "https://coinbase.com", s("TBSWY3DPEHPK3PXY"), t(d(45)), d(45), d(900)),
		login("18", "PayPal", f("folder-3"), s("charlie.brown@gmail.com"), s("Xz5$nKp8!wQt2#Ym"), "https://paypal.com", nil, t(d(800)), d(250), d(1600)),
		login("19", "HMRC Gateway", f("folder-3"), s("472859201"), s("Lq8!mWp4$xNz3#Kt"), "https://www.tax.service.gov.uk", nil, t(d(400)), d(150), d(1200)),
		login("20", "Pension Portal", f("folder-3"), s("OR-28491"), s("Winter2023!"), "https://pension.acme.com", nil, t(d(950)), d(365), d(1100)),

		// Development folder - logins
		login("21", "Docker Hub", f("folder-4"), s("cbrown"), s("Yt7!kLp2$wQx8#Nz"), "https://hub.docker.com", nil, t(d(420)), d(100), d(700)),
		login("22", "npm", f("folder-4"), s("cbrown"), s("Wq4$mKp9!xNt3#Lz"), "https://npmjs.com", s("UBSWY3DPEHPK3PXZ"), t(d(30)), d(30), d(500)),
		login("23", "PyPI", f("folder-4"), s("cbrown"), s("admin1234"), "https://pypi.org", nil, t(d(1100)), d(500), d(1200)),
		login("24", "Terraform Cloud", f("folder-4"), s("charlie.brown@acme.com"), s("Km6!nWp3$wXz8#Qt"), "https://app.terraform.io", s("VBSWY3DPEHPK3PYA"), t(d(20)), d(20), d(400)),
		login("25", "Snyk", f("folder-4"), s("charlie.brown@acme.com"), s("Xp9$kMp5!wNt2#Lq"), "https://app.snyk.io", nil, t(d(300)), d(80), d(500)),
		login("26", "Cloudflare", f("folder-4"), s("charlie.brown@gmail.com"), s("Nz3!mKp7$xQw8#Yt"), "https://dash.cloudflare.com", s("WBSWY3DPEHPK3PYB"), t(d(50)), d(50), d(600)),
		login("27", "DigitalOcean", f("folder-4"), s("charlie.brown@gmail.com"), s("Lt8$nWp4!xKz2#Qm"), "https://cloud.digitalocean.com", nil, t(d(650)), d(200), d(900)),
		login("28", "Vercel", f("folder-4"), s("cbrown"), s("Wt2!kMp6$xNq9#Lz"), "https://vercel.com", nil, t(d(500)), d(150), d(700)),

		// Social folder - logins
		login("29", "LinkedIn", f("folder-5"), s("charlie.brown@gmail.com"), s("Qz7$mKp3!wXt8#Yn"), "https://linkedin.com", nil, t(d(700)), d(250), d(1300)),
		login("30", "Mastodon", f("folder-5"), s("@cbrown@infosec.exchange"), s("Yt5!nLp9$wQk2#Mz"), "https://infosec.exchange", s("XBSWY3DPEHPK3PYC"), t(d(15)), d(15), d(300)),
		login("31", "Discord", f("folder-5"), s("cbrown"), s("Km4$mWp8!xNz3#Lt"), "https://discord.com", nil, t(d(480)), d(160), d(800)),
		login("32", "Strava", f("folder-5"), s("charlie.brown@gmail.com"), s("running2022"), "https://strava.com", nil, t(d(850)), d(350), d(1000)),

		// No folder - logins
		login("33", "Old Forum Account", nil, s("charlie_b2"), s("qwerty123"), "https://forum.example.com", nil, t(d(1500)), d(600), d(2000)),
		login("34", "Random WiFi Portal", nil, s("guest"), s("guest123"), "https://wifi.hotel.com", nil, t(d(200)), d(200), d(200)),

		// Cards
		card("35", "Personal Visa", f("folder-3"), s("Charlie Brown"), s("visa"), s("4111XXXXXXXX1234"), s("09"), s("2027"), d(30), d(800)),
		card("36", "Company Amex", f("folder-1"), s("Charlie Brown"), s("amex"), s("3782XXXXXXX0005"), s("12"), s("2026"), d(60), d(400)),

		// Secure notes
		note("37", "AWS Access Keys (Personal)", f("folder-4"), s("Access Key: AKIAIOSFODNN7EXAMPLE\nSecret Key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n\nRegion: eu-west-1"), d(30), d(300)),
		note("38", "Home WiFi Details", f("folder-2"), s("SSID: BrownHome5G\nPassword: MyW1f1P@ssw0rd!\nRouter admin: 192.168.1.1 / admin / admin"), d(10), d(500)),
		note("39", "Recovery Codes - GitHub", f("folder-4"), s("a1b2c-d3e4f\ng5h6i-j7k8l\nm9n0o-p1q2r\ns3t4u-v5w6x"), d(5), d(400)),
		note("40", "Server SSH Keys", f("folder-1"), s("Production bastion: ssh -i ~/.ssh/prod_rsa charlie@bastion.acme.com\nStaging: ssh -i ~/.ssh/staging_rsa charlie@staging.acme.com"), d(20), d(600)),

		// Identities
		identity("41", "Personal Identity", f("folder-2"), s("Charlie"), s("Brown"), s("charlie.brown@gmail.com"), s("+447700900000"), s(""), d(90), d(1000)),
		identity("42", "Work Identity", f("folder-1"), s("Charlie"), s("Brown"), s("charlie.brown@acme.com"), s("+447700900001"), s("Acme Corp"), d(30), d(500)),
	}
}

func login(id, name string, folderID, username, password *string, uri string, totp *string, pwRevDate *time.Time, revDate, createdDate time.Time) bw.Item {
	return bw.Item{
		ID:           id,
		Type:         bw.ItemTypeLogin,
		Name:         name,
		FolderID:     folderID,
		RevisionDate: revDate,
		CreationDate: createdDate,
		Login: &bw.Login{
			Username:             username,
			Password:             password,
			URIs:                 []bw.URI{{URI: uri}},
			TOTP:                 totp,
			PasswordRevisionDate: pwRevDate,
		},
	}
}

func card(id, name string, folderID, holder, brand, number, expMonth, expYear *string, revDate, createdDate time.Time) bw.Item {
	return bw.Item{
		ID:           id,
		Type:         bw.ItemTypeCard,
		Name:         name,
		FolderID:     folderID,
		RevisionDate: revDate,
		CreationDate: createdDate,
		Card: &bw.Card{
			CardholderName: holder,
			Brand:          brand,
			Number:         number,
			ExpMonth:       expMonth,
			ExpYear:        expYear,
		},
	}
}

func note(id, name string, folderID, notes *string, revDate, createdDate time.Time) bw.Item {
	return bw.Item{
		ID:           id,
		Type:         bw.ItemTypeSecureNote,
		Name:         name,
		FolderID:     folderID,
		Notes:        notes,
		RevisionDate: revDate,
		CreationDate: createdDate,
		SecureNote:   &bw.SecureNote{Type: 0},
	}
}

func identity(id, name string, folderID, firstName, lastName, email, phone, company *string, revDate, createdDate time.Time) bw.Item {
	return bw.Item{
		ID:           id,
		Type:         bw.ItemTypeIdentity,
		Name:         name,
		FolderID:     folderID,
		RevisionDate: revDate,
		CreationDate: createdDate,
		Identity: &bw.Identity{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Phone:     phone,
			Company:   company,
		},
	}
}
