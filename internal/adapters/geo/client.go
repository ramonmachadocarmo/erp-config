package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"erp/services/config-service/internal/domain"
)

var digits = regexp.MustCompile(`\D`)

type Client struct {
	http *http.Client
}

func New() *Client {
	return &Client{http: &http.Client{Timeout: 12 * time.Second}}
}

func (c *Client) ByCEP(ctx context.Context, cep string) (domain.Address, error) {
	cep = digits.ReplaceAllString(cep, "")
	if len(cep) != 8 {
		return domain.Address{}, domain.ErrInvalid
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		return domain.Address{}, err
	}
	var raw struct {
		Erro        flexBool `json:"erro"`
		CEP         string   `json:"cep"`
		Logradouro  string   `json:"logradouro"`
		Complemento string   `json:"complemento"`
		Bairro      string   `json:"bairro"`
		Localidade  string   `json:"localidade"`
		UF          string   `json:"uf"`
	}
	if err := c.decode(req, &raw); err != nil {
		return domain.Address{}, err
	}
	if raw.Erro {
		return domain.Address{}, domain.ErrNotFound
	}
	return domain.Address{
		Zip: raw.CEP, Street: raw.Logradouro, Complement: raw.Complemento,
		District: raw.Bairro, City: raw.Localidade, State: raw.UF,
	}, nil
}

// SearchCEP finds CEPs from UF + city + street via ViaCEP's address search
// (/ws/UF/city/street/json/), which returns up to 50 matches. ViaCEP has no district filter,
// so district is applied here; if it matches nothing (typo, different spelling) the full
// list is returned rather than an empty one.
func (c *Client) SearchCEP(ctx context.Context, state, city, street, district string) ([]domain.Address, error) {
	state = strings.ToUpper(strings.TrimSpace(state))
	city, street = strings.TrimSpace(city), strings.TrimSpace(street)
	if len(state) != 2 || len([]rune(city)) < 3 || len([]rune(street)) < 3 {
		return nil, domain.ErrInvalid
	}
	u := "https://viacep.com.br/ws/" + url.PathEscape(state) + "/" + url.PathEscape(city) + "/" + url.PathEscape(street) + "/json/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	var raw []struct {
		CEP         string `json:"cep"`
		Logradouro  string `json:"logradouro"`
		Complemento string `json:"complemento"`
		Bairro      string `json:"bairro"`
		Localidade  string `json:"localidade"`
		UF          string `json:"uf"`
	}
	if err := c.decode(req, &raw); err != nil {
		return nil, err
	}
	out := make([]domain.Address, 0, len(raw))
	for _, r := range raw {
		out = append(out, domain.Address{
			Zip: r.CEP, Street: r.Logradouro, Complement: r.Complemento,
			District: r.Bairro, City: r.Localidade, State: r.UF,
		})
	}
	if len(out) == 0 {
		return nil, domain.ErrNotFound
	}
	return filterDistrict(out, district), nil
}

func filterDistrict(list []domain.Address, district string) []domain.Address {
	d := fold(district)
	if d == "" {
		return list
	}
	var out []domain.Address
	for _, a := range list {
		if strings.Contains(fold(a.District), d) {
			out = append(out, a)
		}
	}
	if len(out) == 0 {
		return list
	}
	return out
}

func (c *Client) ByGeo(ctx context.Context, lat, lng float64) (domain.Address, error) {
	q := url.Values{}
	q.Set("format", "jsonv2")
	q.Set("lat", fmt.Sprintf("%f", lat))
	q.Set("lon", fmt.Sprintf("%f", lng))
	q.Set("addressdetails", "1")
	q.Set("accept-language", "pt-BR")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://nominatim.openstreetmap.org/reverse?"+q.Encode(), nil)
	if err != nil {
		return domain.Address{}, err
	}
	req.Header.Set("User-Agent", "erp-config-service/1.0")
	var raw struct {
		Address struct {
			Road          string `json:"road"`
			HouseNumber   string `json:"house_number"`
			Suburb        string `json:"suburb"`
			Neighbourhood string `json:"neighbourhood"`
			City          string `json:"city"`
			Town          string `json:"town"`
			Village       string `json:"village"`
			State         string `json:"state"`
			ISO           string `json:"ISO3166-2-lvl4"`
			Postcode      string `json:"postcode"`
		} `json:"address"`
	}
	if err := c.decode(req, &raw); err != nil {
		return domain.Address{}, err
	}
	city := first(raw.Address.City, raw.Address.Town, raw.Address.Village)
	state := raw.Address.ISO
	if i := strings.LastIndex(state, "-"); i >= 0 {
		state = state[i+1:]
	}
	latc, lngc := lat, lng
	return domain.Address{
		Zip: raw.Address.Postcode, Street: raw.Address.Road, Number: raw.Address.HouseNumber,
		District: first(raw.Address.Suburb, raw.Address.Neighbourhood),
		City:     city, State: state, Lat: &latc, Lng: &lngc,
	}, nil
}

func (c *Client) Search(ctx context.Context, query string) (domain.Address, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return domain.Address{}, domain.ErrInvalid
	}
	q := url.Values{}
	q.Set("format", "jsonv2")
	q.Set("q", query)
	q.Set("limit", "1")
	q.Set("addressdetails", "1")
	q.Set("countrycodes", "br")
	q.Set("accept-language", "pt-BR")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://nominatim.openstreetmap.org/search?"+q.Encode(), nil)
	if err != nil {
		return domain.Address{}, err
	}
	req.Header.Set("User-Agent", "erp-config-service/1.0")
	var raw []struct {
		Lat     string `json:"lat"`
		Lon     string `json:"lon"`
		Address struct {
			Road          string `json:"road"`
			HouseNumber   string `json:"house_number"`
			Suburb        string `json:"suburb"`
			Neighbourhood string `json:"neighbourhood"`
			City          string `json:"city"`
			Town          string `json:"town"`
			Village       string `json:"village"`
			State         string `json:"state"`
			ISO           string `json:"ISO3166-2-lvl4"`
			Postcode      string `json:"postcode"`
		} `json:"address"`
	}
	if err := c.decode(req, &raw); err != nil {
		return domain.Address{}, err
	}
	if len(raw) == 0 {
		return domain.Address{}, domain.ErrNotFound
	}
	hit := raw[0]
	var lat, lng float64
	fmt.Sscanf(hit.Lat, "%f", &lat)
	fmt.Sscanf(hit.Lon, "%f", &lng)
	if lat == 0 && lng == 0 {
		return domain.Address{}, domain.ErrNotFound
	}
	city := first(hit.Address.City, hit.Address.Town, hit.Address.Village)
	state := hit.Address.ISO
	if i := strings.LastIndex(state, "-"); i >= 0 {
		state = state[i+1:]
	}
	return domain.Address{
		Zip: hit.Address.Postcode, Street: hit.Address.Road, Number: hit.Address.HouseNumber,
		District: first(hit.Address.Suburb, hit.Address.Neighbourhood),
		City: city, State: state, Lat: &lat, Lng: &lng,
	}, nil
}

type nominatimHit struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Addresstype string `json:"addresstype"`
	Name        string `json:"name"`
	Address     struct {
		Road          string `json:"road"`
		HouseNumber   string `json:"house_number"`
		Suburb        string `json:"suburb"`
		Neighbourhood string `json:"neighbourhood"`
		City          string `json:"city"`
		Town          string `json:"town"`
		Village       string `json:"village"`
		Municipality  string `json:"municipality"`
		State         string `json:"state"`
		ISO           string `json:"ISO3166-2-lvl4"`
		Postcode      string `json:"postcode"`
	} `json:"address"`
}

func (c *Client) SearchPlace(ctx context.Context, street, number, district, city, state, zip string) (domain.Address, error) {
	street, number, district = strings.TrimSpace(street), strings.TrimSpace(number), strings.TrimSpace(district)
	city, state, zip = strings.TrimSpace(city), strings.ToUpper(strings.TrimSpace(state)), digits.ReplaceAllString(zip, "")
	line := street
	if number != "" {
		line = strings.TrimSpace(number + " " + street)
	}
	stateName := ufName[state]
	if stateName == "" {
		stateName = state
	}
	attempts := []url.Values{}
	for _, parts := range [][]string{
		{line, district, city, stateName, "Brasil"},
		{street, district, city, stateName, "Brasil"},
		{district, city, stateName, "Brasil"},
		{city, stateName, "Brasil"},
	} {
		q := joinQuery(parts)
		if q == "" {
			continue
		}
		if len(attempts) > 0 && attempts[len(attempts)-1].Get("q") == q {
			continue
		}
		p := nominatimBase()
		p.Set("q", q)
		attempts = append(attempts, p)
	}
	for i, vals := range attempts {
		if i > 0 {
			select {
			case <-ctx.Done():
				return domain.Address{}, ctx.Err()
			case <-time.After(1100 * time.Millisecond):
			}
		}
		hits, err := c.nominatimSearch(ctx, vals)
		if err != nil || len(hits) == 0 {
			continue
		}
		a, err := pickBest(hits, city, state, zip, district, street)
		if err != nil {
			continue
		}
		if street != "" && a.Street == "" {
			a.Street = street
		}
		if number != "" && a.Number == "" {
			a.Number = number
		}
		if district != "" && a.District == "" {
			a.District = district
		}
		if city != "" && a.City == "" {
			a.City = city
		}
		if state != "" && a.State == "" {
			a.State = state
		}
		if zip != "" && a.Zip == "" {
			a.Zip = zip
		}
		return a, nil
	}
	if district == "" {
		if lat, lng, ok := cityFallback(city, state); ok {
			return domain.Address{Street: street, Number: number, District: district, City: city, State: state, Zip: zip, Lat: &lat, Lng: &lng}, nil
		}
	}
	return domain.Address{}, domain.ErrNotFound
}

func joinQuery(parts []string) string {
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, x := range parts {
		x = strings.TrimSpace(x)
		k := fold(x)
		if x == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, x)
	}
	if len(out) < 2 {
		return ""
	}
	return strings.Join(out, ", ")
}

func nominatimBase() url.Values {
	q := url.Values{}
	q.Set("format", "jsonv2")
	q.Set("limit", "5")
	q.Set("addressdetails", "1")
	q.Set("countrycodes", "br")
	q.Set("accept-language", "pt-BR")
	return q
}

func (c *Client) nominatimSearch(ctx context.Context, q url.Values) ([]nominatimHit, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://nominatim.openstreetmap.org/search?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ERP-Logistica/1.0 (rotas)")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 || len(body) == 0 || body[0] != '[' {
		return nil, nil
	}
	var raw []nominatimHit
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, nil
	}
	return raw, nil
}

func pickBest(hits []nominatimHit, city, state, zip, district, street string) (domain.Address, error) {
	bestScore := -1
	var best nominatimHit
	for _, h := range hits {
		sc := scoreHit(h, city, state, zip, district, street)
		if sc > bestScore {
			bestScore = sc
			best = h
		}
	}
	need := 0
	if city != "" || state != "" {
		need = 2
	}
	if district != "" {
		need += 4
	}
	if bestScore < need {
		return domain.Address{}, domain.ErrNotFound
	}
	return hitAddress(best)
}

func scoreHit(h nominatimHit, city, state, zip, district, street string) int {
	lat, lng := 0.0, 0.0
	fmt.Sscanf(h.Lat, "%f", &lat)
	fmt.Sscanf(h.Lon, "%f", &lng)
	if city != "" && state != "" && !inUF(state, lat, lng) {
		return -1
	}
	kind := strings.ToLower(h.Addresstype)
	if district != "" && (kind == "city" || kind == "state" || kind == "country" || kind == "municipality") {
		return -1
	}
	score := 0
	gotCity := fold(first(h.Address.City, h.Address.Town, h.Address.Village, h.Address.Municipality))
	wantCity := fold(city)
	if wantCity != "" && gotCity != "" && (gotCity == wantCity || strings.Contains(gotCity, wantCity) || strings.Contains(wantCity, gotCity)) {
		score += 3
	}
	uf := strings.ToUpper(strings.TrimSpace(state))
	iso := strings.ToUpper(h.Address.ISO)
	gotState := fold(h.Address.State)
	if uf != "" && (strings.HasSuffix(iso, "-"+uf) || gotState == fold(ufName[uf]) || gotState == fold(uf) || inUF(uf, lat, lng)) {
		score += 2
	}
	wantDist := fold(district)
	gotDist := fold(first(h.Address.Suburb, h.Address.Neighbourhood, h.Name))
	if wantDist != "" && gotDist != "" && (gotDist == wantDist || strings.Contains(gotDist, wantDist) || strings.Contains(wantDist, gotDist)) {
		score += 5
	}
	wantStreet := fold(street)
	gotStreet := fold(h.Address.Road)
	if wantStreet != "" && gotStreet != "" && (gotStreet == wantStreet || strings.Contains(gotStreet, wantStreet) || strings.Contains(wantStreet, gotStreet)) {
		score += 3
	}
	wantZip := digits.ReplaceAllString(zip, "")
	gotZip := digits.ReplaceAllString(h.Address.Postcode, "")
	if len(wantZip) >= 5 && len(gotZip) >= 5 && wantZip[:5] == gotZip[:5] {
		score += 2
	}
	return score
}

func hitAddress(hit nominatimHit) (domain.Address, error) {
	var lat, lng float64
	fmt.Sscanf(hit.Lat, "%f", &lat)
	fmt.Sscanf(hit.Lon, "%f", &lng)
	if lat == 0 && lng == 0 {
		return domain.Address{}, domain.ErrNotFound
	}
	city := first(hit.Address.City, hit.Address.Town, hit.Address.Village, hit.Address.Municipality)
	state := hit.Address.ISO
	if i := strings.LastIndex(state, "-"); i >= 0 {
		state = state[i+1:]
	}
	return domain.Address{
		Zip: hit.Address.Postcode, Street: hit.Address.Road, Number: hit.Address.HouseNumber,
		District: first(hit.Address.Suburb, hit.Address.Neighbourhood),
		City: city, State: state, Lat: &lat, Lng: &lng,
	}, nil
}

func cityFallback(city, uf string) (lat, lng float64, ok bool) {
	key := fold(city) + "|" + strings.ToUpper(strings.TrimSpace(uf))
	p, ok := cityCenter[key]
	if !ok {
		return 0, 0, false
	}
	return p[0], p[1], true
}

var cityCenter = map[string][2]float64{
	"rio branco|AC": {-9.974, -67.824}, "maceio|AL": {-9.666, -35.735}, "macapa|AP": {0.034, -51.069},
	"manaus|AM": {-3.119, -60.022}, "salvador|BA": {-12.971, -38.501}, "fortaleza|CE": {-3.732, -38.527},
	"brasilia|DF": {-15.794, -47.882}, "vitoria|ES": {-20.315, -40.312}, "goiania|GO": {-16.686, -49.264},
	"sao luis|MA": {-2.530, -44.307}, "cuiaba|MT": {-15.601, -56.097}, "campo grande|MS": {-20.469, -54.620},
	"belo horizonte|MG": {-19.917, -43.935}, "belem|PA": {-1.456, -48.504}, "joao pessoa|PB": {-7.115, -34.863},
	"curitiba|PR": {-25.429, -49.272}, "recife|PE": {-8.054, -34.881}, "teresina|PI": {-5.089, -42.802},
	"rio de janeiro|RJ": {-22.907, -43.173}, "natal|RN": {-5.794, -35.211}, "porto alegre|RS": {-30.034, -51.218},
	"porto velho|RO": {-8.761, -63.900}, "boa vista|RR": {2.823, -60.676}, "florianopolis|SC": {-27.595, -48.548},
	"sao paulo|SP": {-23.551, -46.633}, "aracaju|SE": {-10.909, -37.075}, "palmas|TO": {-10.184, -48.334},
}

func fold(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	r := strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i", "ó", "o", "ô", "o", "õ", "o", "ú", "u", "ç", "c")
	return r.Replace(s)
}

func inUF(uf string, lat, lng float64) bool {
	b, ok := ufBox[strings.ToUpper(strings.TrimSpace(uf))]
	if !ok {
		return lat >= -34 && lat <= 6 && lng >= -74 && lng <= -32
	}
	return lat >= b[0] && lat <= b[1] && lng >= b[2] && lng <= b[3]
}

var ufName = map[string]string{
	"AC": "Acre", "AL": "Alagoas", "AP": "Amapá", "AM": "Amazonas", "BA": "Bahia",
	"CE": "Ceará", "DF": "Distrito Federal", "ES": "Espírito Santo", "GO": "Goiás",
	"MA": "Maranhão", "MT": "Mato Grosso", "MS": "Mato Grosso do Sul", "MG": "Minas Gerais",
	"PA": "Pará", "PB": "Paraíba", "PR": "Paraná", "PE": "Pernambuco", "PI": "Piauí",
	"RJ": "Rio de Janeiro", "RN": "Rio Grande do Norte", "RS": "Rio Grande do Sul",
	"RO": "Rondônia", "RR": "Roraima", "SC": "Santa Catarina", "SP": "São Paulo",
	"SE": "Sergipe", "TO": "Tocantins",
}

var ufBox = map[string][4]float64{
	"AC": {-11.2, -7.1, -74.0, -66.6}, "AL": {-10.5, -8.8, -38.3, -35.1},
	"AP": {-1.3, 4.5, -54.9, -49.8}, "AM": {-11.2, 2.3, -73.9, -56.0},
	"BA": {-18.4, -8.5, -46.6, -37.3}, "CE": {-7.9, -2.8, -41.5, -37.2},
	"DF": {-16.1, -15.4, -48.3, -47.3}, "ES": {-21.3, -17.9, -41.9, -28.8},
	"GO": {-19.5, -12.4, -53.3, -45.9}, "MA": {-10.3, -1.0, -48.8, -41.8},
	"MT": {-18.1, -7.3, -61.7, -50.2}, "MS": {-24.1, -17.1, -58.2, -50.9},
	"MG": {-23.0, -14.2, -51.1, -39.8}, "PA": {-9.9, 2.6, -58.9, -46.0},
	"PB": {-8.4, -6.0, -38.9, -34.7}, "PR": {-26.8, -22.5, -54.7, -48.0},
	"PE": {-9.6, -7.1, -41.4, -34.8}, "PI": {-10.9, -2.7, -45.9, -40.3},
	"RJ": {-23.4, -20.7, -44.9, -40.9}, "RN": {-7.0, -4.8, -38.6, -34.9},
	"RS": {-33.8, -27.0, -57.7, -49.6}, "RO": {-13.7, -7.9, -66.9, -59.7},
	"RR": {-1.6, 5.3, -64.8, -58.8}, "SC": {-29.4, -25.9, -53.9, -48.3},
	"SP": {-25.4, -19.7, -53.2, -44.1}, "SE": {-11.6, -9.5, -38.3, -36.3},
	"TO": {-13.5, -5.1, -50.8, -45.6},
}

func (c *Client) decode(req *http.Request, dest any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("geo: %s", strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, dest)
}

func first(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

type flexBool bool

func (b *flexBool) UnmarshalJSON(p []byte) error {
	s := strings.Trim(strings.TrimSpace(string(p)), `"`)
	*b = s == "true" || s == "1"
	return nil
}
