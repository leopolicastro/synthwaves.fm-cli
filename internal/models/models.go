package models

// ArtistRef is a minimal artist reference embedded in other resources.
type ArtistRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Artist represents a full artist resource.
type Artist struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
	AlbumsCount int    `json:"albums_count"`
	TracksCount int    `json:"tracks_count"`
	CreatedAt   string `json:"created_at"`
}

// ArtistDetail is an artist with their albums (from GET /artists/:id).
type ArtistDetail struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Category    string         `json:"category"`
	ImageURL    string         `json:"image_url"`
	AlbumsCount int            `json:"albums_count"`
	TracksCount int            `json:"tracks_count"`
	CreatedAt   string         `json:"created_at"`
	Albums      []AlbumSummary `json:"albums"`
}

// AlbumRef is a minimal album reference embedded in other resources.
type AlbumRef struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// AlbumSummary is used in artist detail and search results.
type AlbumSummary struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Year          int    `json:"year"`
	Genre         string `json:"genre"`
	TracksCount   int    `json:"tracks_count"`
	CoverImageURL string `json:"cover_image_url"`
}

// Album is the full album resource from list endpoints.
type Album struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Year          int       `json:"year"`
	Genre         string    `json:"genre"`
	TracksCount   int       `json:"tracks_count"`
	CoverImageURL string    `json:"cover_image_url"`
	Artist        ArtistRef `json:"artist"`
	CreatedAt     string    `json:"created_at"`
}

// AlbumDetail is an album with its tracks (from GET /albums/:id).
type AlbumDetail struct {
	ID            int64        `json:"id"`
	Title         string       `json:"title"`
	Year          int          `json:"year"`
	Genre         string       `json:"genre"`
	TracksCount   int          `json:"tracks_count"`
	CoverImageURL string       `json:"cover_image_url"`
	Artist        ArtistRef    `json:"artist"`
	CreatedAt     string       `json:"created_at"`
	TotalDuration float64      `json:"total_duration"`
	Tracks        []TrackBrief `json:"tracks"`
}

// TrackBrief is a minimal track used in album detail and playlist listings.
type TrackBrief struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	TrackNumber int     `json:"track_number"`
	DiscNumber  int     `json:"disc_number"`
	Duration    float64 `json:"duration"`
	FileFormat  string  `json:"file_format"`
	HasAudio    bool    `json:"has_audio"`
}

// TrackEmbedded is used in search results and play history.
type TrackEmbedded struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	Duration float64   `json:"duration"`
	Artist   ArtistRef `json:"artist"`
	Album    AlbumRef  `json:"album"`
}

// Track is the full track resource from list endpoints.
type Track struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	TrackNumber int       `json:"track_number"`
	DiscNumber  int       `json:"disc_number"`
	Duration    float64   `json:"duration"`
	Bitrate     int       `json:"bitrate"`
	FileFormat  string    `json:"file_format"`
	FileSize    int64     `json:"file_size"`
	Lyrics      string    `json:"lyrics"`
	HasAudio    bool      `json:"has_audio"`
	Artist      ArtistRef `json:"artist"`
	Album       AlbumRef  `json:"album"`
	CreatedAt   string    `json:"created_at"`
}

// StreamInfo is the response from GET /tracks/:id/stream.
type StreamInfo struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
}

// PlaylistRef is a minimal playlist reference.
type PlaylistRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Playlist is the full playlist resource from list endpoints.
type Playlist struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	TracksCount int    `json:"tracks_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// PlaylistDetail is a playlist with its tracks (from GET /playlists/:id).
type PlaylistDetail struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	TracksCount   int             `json:"tracks_count"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
	TotalDuration float64         `json:"total_duration"`
	Tracks        []PlaylistTrack `json:"tracks"`
	Pagination    Pagination      `json:"pagination"`
}

// PlaylistTrack is a track within a playlist, with position info.
type PlaylistTrack struct {
	PlaylistTrackID int64         `json:"playlist_track_id"`
	Position        int           `json:"position"`
	Track           TrackEmbedded `json:"track"`
}

// PlaylistTrackResult is the response from adding tracks to a playlist.
type PlaylistTrackResult struct {
	Added       int `json:"added"`
	TracksCount int `json:"tracks_count"`
}

// Favorite represents a favorited resource.
type Favorite struct {
	ID            int64           `json:"id"`
	FavorableType string          `json:"favorable_type"`
	FavorableID   int64           `json:"favorable_id"`
	CreatedAt     string          `json:"created_at"`
	Favorable     FavorableObject `json:"favorable"`
}

// FavorableObject is the polymorphic nested object in a Favorite.
// For Tracks/Albums it has Title + Artist, for Artists it has Name.
type FavorableObject struct {
	ID     int64     `json:"id"`
	Title  string    `json:"title"`
	Name   string    `json:"name"`
	Artist *struct {
		Name string `json:"name"`
	} `json:"artist"`
}

// DisplayName returns a human-readable name for the favorited item.
func (f FavorableObject) DisplayName() string {
	if f.Title != "" {
		return f.Title
	}
	return f.Name
}

// ArtistName returns the artist name if available.
func (f FavorableObject) ArtistName() string {
	if f.Artist != nil {
		return f.Artist.Name
	}
	return ""
}

// PlayHistory is a single play event.
type PlayHistory struct {
	ID       int64         `json:"id"`
	Track    TrackEmbedded `json:"track"`
	PlayedAt string        `json:"played_at"`
}

// RadioStation represents a personal radio station.
type RadioStation struct {
	ID                int64        `json:"id"`
	Name              string       `json:"name"`
	Status            string       `json:"status"`
	MountPoint        string       `json:"mount_point"`
	PlaybackMode      string       `json:"playback_mode"`
	Bitrate           int          `json:"bitrate"`
	CrossfadeDuration float64      `json:"crossfade_duration"`
	ListenURL         string       `json:"listen_url"`
	Playlist          PlaylistRef  `json:"playlist"`
	CurrentTrack      *TrackMinimal `json:"current_track"`
	CreatedAt         string       `json:"created_at"`
}

// PublicRadioStation represents a public/community radio station.
type PublicRadioStation struct {
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	Status        string        `json:"status"`
	Slug          string        `json:"slug"`
	ListenURL     string        `json:"listen_url"`
	ListenerCount int           `json:"listener_count"`
	ImageURL      string        `json:"image_url"`
	CurrentTrack  *TrackMinimal `json:"current_track"`
}

// TrackMinimal is used in radio station current track.
type TrackMinimal struct {
	ID     int64     `json:"id"`
	Title  string    `json:"title"`
	Artist ArtistRef `json:"artist"`
}

// RadioControlResult is the response from radio station control actions.
type RadioControlResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Tag represents a tag (genre, mood, etc.).
type Tag struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	TagType string `json:"tag_type"`
}

// Tagging represents a tag-to-resource association.
type Tagging struct {
	ID           int64  `json:"id"`
	Tag          Tag    `json:"tag"`
	TaggableType string `json:"taggable_type"`
	TaggableID   int64  `json:"taggable_id"`
}

// Profile is the current user's profile.
type Profile struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	EmailAddress string       `json:"email_address"`
	Theme        string       `json:"theme"`
	CreatedAt    string       `json:"created_at"`
	Stats        ProfileStats `json:"stats"`
}

// ProfileStats are the user's library counts.
type ProfileStats struct {
	ArtistsCount   int `json:"artists_count"`
	AlbumsCount    int `json:"albums_count"`
	TracksCount    int `json:"tracks_count"`
	PlaylistsCount int `json:"playlists_count"`
}

// SearchResult holds cross-resource search results.
type SearchResult struct {
	Artists []ArtistRef         `json:"artists"`
	Albums  []AlbumSearchResult `json:"albums"`
	Tracks  []TrackEmbedded     `json:"tracks"`
}

// AlbumSearchResult is the album shape returned by the search endpoint.
type AlbumSearchResult struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Year          int       `json:"year"`
	Genre         string    `json:"genre"`
	CoverImageURL string    `json:"cover_image_url"`
	Artist        ArtistRef `json:"artist"`
}

// TrackMetadata holds available filter values.
type TrackMetadata struct {
	Genres    []string `json:"genres"`
	Languages []string `json:"languages"`
	Decades   []string `json:"decades"`
}

// Stats holds library and listening statistics.
type Stats struct {
	Library   LibraryStats   `json:"library"`
	Listening ListeningStats `json:"listening"`
}

// LibraryStats are aggregate library counts.
type LibraryStats struct {
	ArtistsCount   int     `json:"artists_count"`
	AlbumsCount    int     `json:"albums_count"`
	TracksCount    int     `json:"tracks_count"`
	TotalDuration  float64 `json:"total_duration"`
	TotalFileSize  int64   `json:"total_file_size"`
	PlaylistsCount int     `json:"playlists_count"`
}

// ListeningStats are time-ranged listening analytics.
type ListeningStats struct {
	TimeRange          string         `json:"time_range"`
	TotalPlays         int            `json:"total_plays"`
	TotalListeningTime float64        `json:"total_listening_time"`
	CurrentStreak      int            `json:"current_streak"`
	LongestStreak      int            `json:"longest_streak"`
	TopTracks          []TopTrack     `json:"top_tracks"`
	TopArtists         []TopArtist    `json:"top_artists"`
	TopGenres          []TopGenre     `json:"top_genres"`
	HourlyDistribution []int `json:"hourly_distribution"`
	DailyDistribution  []int `json:"daily_distribution"`
}

type TopTrack struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	ArtistName string `json:"artist_name"`
	PlayCount  int    `json:"play_count"`
}

type TopArtist struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	PlayCount int    `json:"play_count"`
}

type TopGenre struct {
	Genre     string `json:"genre"`
	PlayCount int    `json:"play_count"`
}

// AuthToken is the response from POST /auth/token.
type AuthToken struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}
