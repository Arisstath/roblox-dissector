package peer

func ExampleNewCustomClient() {
	client := NewCustomClient()
	// The following values have been dumped from a 0x8A packet
	client.SecurityKey = "571cb33a3b024d7b8dafb87156909e92b7eaf86d!1ac9a51ce47836b5c1f65dfc441dfa41"
	client.OsPlatform = "Win32"
	client.GoldenHash = 19857408
	client.DataModelHash = "4b8387d8b57d73944b33dbe044b3707b"

	// The BrowserTrackerId can be dumped from the Roblox site
	client.BrowserTrackerId = 9783257674
	
	// Connects a Guest user using 12109643 (Fencing) as the place id and 2 (male) 
	// as the gender id
	err := client.ConnectGuest(12109643, 2)
	if err != nil {
		panic(err)
	}
}
