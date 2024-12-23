package bd

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB DBentity

// ---------------------------------------->>>INITIALIZATION---------------------------------------------------------------------
func Init(host, user, password, dbname string, port int, sslmode string) (err error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", host, user, password, dbname, port, sslmode)
	DB.Socket, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		err = fmt.Errorf("database init error: %w", err)
		return
	}
	return nil
}

func Migrate() (err error) {
	if err = DB.Socket.AutoMigrate(UserData{}, JobAnnounce{}, UserPivotVacancy{}, CountrySQL{}, Region{}, City{}, Schedule{}, VacancynameSearchPattern{}); err != nil {
		err = fmt.Errorf("database automigration error: %w", err)
	}
	return
}

// ----------------------------------------<<<INITIALIZATION----------------------------------------------------------------------

// ------------------------------------------------------------->>>LOCATION WRITERS-----------------------------------------------------
func (countries SQLcountries) WriteToDB() (err error) {
	if err = DB.Socket.Save(&countries).Error; err != nil {
		err = fmt.Errorf("list of region save error: %w", err)
	}
	return
}

func (regions SQLregions) WriteToDB() (err error) {
	if err = DB.Socket.Save(&regions).Error; err != nil {
		err = fmt.Errorf("list of region save error: %w", err)
	}
	return
}

func (cities SQLcities) WriteToDB() (err error) {
	if err = DB.Socket.Save(&cities).Error; err != nil {
		err = fmt.Errorf("list of region save error: %w", err)
	}
	return
}

// -------------------------------------------------------------<<<LOCATION WRITERS-----------------------------------------------------
// S--U--
func FindOrCreateUser(tgID int64) (u UserData, err error) {
	if err = DB.Socket.Where("tg_id=?", tgID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.TgID = tgID
			u.Schedule = "fullDay"
			if err = DB.Socket.Create(&u).Error; err != nil {
				err = fmt.Errorf("user creating error: %w", err)
			}
			return
		} else {
			err = fmt.Errorf("user finding error: %w", err)
			return
		}
	}
	return
}

// User Data Update
func (u UserData) Update() (err error) {
	sqlu, err := FindOrCreateUser(u.TgID)
	if err != nil {
		err = fmt.Errorf("update user data error: %w", err)
		return
	}
	sqlu.ExperienceYear = u.ExperienceYear
	sqlu.Schedule = u.Schedule
	sqlu.VacancyName = u.VacancyName
	if err = DB.Socket.Model(&sqlu).Select("vacancy_name", "experience_year").Updates(UserData{VacancyName: sqlu.VacancyName, ExperienceYear: sqlu.ExperienceYear}).Error; err != nil {
		err = fmt.Errorf("updating userData error: %w", err)
		return
	}

	WorkDue <- true

	return nil
}

func (u UserData) UpdateLocation() (err error) {
	if err = DB.Socket.Model(&u).Where("tg_id=?", u.TgID).Update("location", u.Location).Error; err != nil {
		err = fmt.Errorf("user data location on db update error: %w", err)
		return
	}

	WorkDue <- true

	return nil
}

func (u UserData) UpdateSchedule() (err error) {
	if err = DB.Socket.Model(&u).Where("tg_id=?", u.TgID).Update("schedule", u.Schedule).Error; err != nil {
		err = fmt.Errorf("user data schedule field in db update error:%w", err)
		return
	}

	WorkDue <- true

	return nil
}

func (areas SQLcountries) IdsSequence() (iDs []uint) {
	iDs = make([]uint, 0, len(areas))
	for _, area := range areas {
		iDs = append(iDs, area.ID)
	}
	return
}

func (areas SQLregions) IdsSequence() (iDs []uint) {
	iDs = make([]uint, 0, len(areas))
	for _, area := range areas {
		iDs = append(iDs, area.ID)
	}
	return
}

func (areas SQLcities) IdsSequence() (iDs []uint) {
	iDs = make([]uint, 0, len(areas))
	for _, area := range areas {
		iDs = append(iDs, area.ID)
	}
	return
}

// idace******DB
func CountriesLis() (areaData Countries, err error) {
	dbSQLCountries := SQLcountries{}
	if err = DB.Socket.Find(&dbSQLCountries).Error; err != nil {
		err = fmt.Errorf("bd CountriesLis getting error: %w", err)
		return
	}

	dbSQLRegions := SQLregions{}
	if err = DB.Socket.Where("owner IN ?", dbSQLCountries.IdsSequence()).Find(&dbSQLRegions).Error; err != nil {
		err = fmt.Errorf("bd Countrieslis: regions getting error: %w", err)
		return
	}

	//to cities region need

	dbSQLCities := SQLcities{}
	if err = DB.Socket.Where("owner IN ?", dbSQLRegions.IdsSequence()).Find(&dbSQLCities).Error; err != nil {
		err = fmt.Errorf("bd CountriesLis: cities getting error: %w", err)
		return
	}

	for _, country := range dbSQLCountries {
		co := CountrieModel{Count: AreaEntity{ID: country.ID, Name: country.Name}}

		for _, region := range dbSQLRegions {
			reg := RegionModel{Region: AreaEntity{ID: region.ID, Name: region.Name, Owner: region.Owner}}

			for _, city := range dbSQLCities {
				if city.Owner == region.ID {
					c := AreaEntity{ID: city.ID, Name: city.Name, Owner: city.Owner}
					reg.Cities = append(reg.Cities, c)
				}
			}

			if region.Owner == country.ID {
				co.Regions = append(co.Regions, reg)
			}
		}
		areaData = append(areaData, co)
	}

	return
}

/*func an() {
	regions := Regions{}
	if err = DB.Socket.Where("owner=?", c.ID).Find(&regions).Error; err != nil {
		continue
	}

	for _, r := range regions {
		regionsData := RegionData{}
		regionsData.Region.ID = r.ID
		regionsData.Region.Name = r.Name
		regionsData.Region.Owner = countrieData.Count.ID

		if err = DB.Socket.Where("owner=?", r.ID).Find(&regionsData.Cities).Error; err != nil {
			continue
		}

		countrieData.Regions = append(countrieData.Regions, regionsData)
	}

	ad.Countries = append(ad.Countries, countrieData)
}*/

func FindCitiesByName(cityName string) (cities SQLcities, err error) {
	if err = DB.Socket.Where("LOWER(name) like ?", "%"+strings.ToLower(cityName)+"%").Find(&cities).Error; err != nil {
		err = fmt.Errorf("cities by name finding error: %w", err)
		return
	}
	IDs := make([]uint, 0, len(cities))
	for _, city := range cities {
		IDs = append(IDs, city.Owner)
	}

	regions := SQLregions{}
	if err = DB.Socket.Where("id in ?", IDs).Find(&regions).Error; err != nil {
		err = fmt.Errorf("regions by id finding error: %w", err)
		return
	}
	IDs = make([]uint, 0, len(regions))
	for _, region := range regions {
		IDs = append(IDs, region.Owner)
	}

	countries := SQLcountries{}
	if err = DB.Socket.Where("id in ?", IDs).Find(&countries).Error; err != nil {
		err = fmt.Errorf("countries by id finding error: %w", err)
		return
	}

	for i, city := range cities {
		for _, region := range regions {
			if city.Owner == region.ID {
				cities[i].Name = region.Name + ", " + city.Name
				for _, country := range countries {
					if region.Owner == country.ID {
						cities[i].Name = country.Name + ", " + cities[i].Name
					}
				}
			}
		}
	}
	return
}

func FindRegionByName(regionName string) (regions SQLregions, err error) {
	if err = DB.Socket.Where("LOWER(name) like ?", "%"+strings.ToLower(regionName)+"%").Find(&regions).Error; err != nil {
		err = fmt.Errorf("Find region by name error: %w", err)
		return
	}
	IDs := make([]uint, 0, len(regions))
	for _, region := range regions {
		IDs = append(IDs, region.Owner)
	}

	countries := SQLcountries{}
	if err = DB.Socket.Where("id in ?", IDs).Find(&countries).Error; err != nil {
		err = fmt.Errorf("find countries by id error: %w", err)
		return
	}

	for i, region := range regions {
		for _, country := range countries {
			if region.Owner == country.ID {
				regions[i].Name = country.Name + ", " + region.Name
			}
		}
	}
	return
}

func FindCountries() (countries SQLcountries, err error) {
	if err = DB.Socket.Find(&countries).Error; err != nil {
		err = fmt.Errorf("countries finding error: %w", err)
	}
	return
}

func (ad Countries) FindLocationByAreaID(areaID int) (country *AreaEntity, region *AreaEntity, city *AreaEntity) {

	for _, countr := range ad {
		if countr.Count.ID == uint(areaID) {
			country = &countr.Count
			return
		}

		for _, reg := range countr.Regions {
			if reg.Region.ID == uint(areaID) {
				country = &countr.Count
				region = &reg.Region
				return
			}

			for _, cit := range reg.Cities {
				if cit.ID == uint(areaID) {
					city = &cit
					region = &reg.Region
					country = &countr.Count
					return
				}
			}
		}
	}

	return
}

func (ad Countries) FindContainLocationIDsList(areaID uint) (locationListIDs []uint) {
	if areaID != 0 {
		for _, country := range ad {

			if country.Count.ID == areaID {
				locationListIDs = append(locationListIDs, country.Count.ID)
				for _, region := range country.Regions {
					locationListIDs = append(locationListIDs, region.Region.ID)
					for _, city := range region.Cities {
						locationListIDs = append(locationListIDs, city.ID)
					}
				}
				return
			}

			for _, region := range country.Regions {
				if region.Region.ID == areaID {
					locationListIDs = append(locationListIDs, region.Region.ID)
					for _, city := range region.Cities {
						locationListIDs = append(locationListIDs, city.ID)
					}
					return
				}

				for _, city := range region.Cities {
					if city.ID == areaID {
						locationListIDs = append(locationListIDs, city.ID)
						return
					}
				}
			}

		}
	}
	return
}

// Поиск локации по ИД
// Проверяет ИД по порядку в таблицах: стран, регионов, населенных пунктов
func FindLocByID(locID uint) (locName string, err error) {
	locName = "не имеет значения"
	country := CountrySQL{}
	if err = DB.Socket.Where("id=?", locID).First(&country).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			region := Region{}
			if err = DB.Socket.Where("id=?", locID).First(&region).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					city := City{}
					if err = DB.Socket.Where("id=?", locID).First(&city).Error; err != nil {
						if !errors.Is(err, gorm.ErrRecordNotFound) {
							err = fmt.Errorf("location by ID finding error: %w", err)
							return
						}
					} else {
						locName = city.Name
					}
				} else {
					err = fmt.Errorf("location by ID finding error: %w", err)
					return
				}

			} else {
				locName = region.Name
			}

		} else {
			err = fmt.Errorf("location by ID finding error: %w", err)
			return
		}

	} else {
		locName = country.Name
	}
	return
}

// schedules list  in DB create or update
func (sch Schedules) CreateToDB() (err error) {
	if err = DB.Socket.Save(&sch).Error; err != nil {
		err = fmt.Errorf("schedule of vacancie annonce create error: %w", err)
	}
	return
}

// shedules finding
func GetSchedule(scheduleID string) (schdules Schedules, err error) {
	if len(scheduleID) == 0 {
		if err = DB.Socket.Find(&schdules).Error; err != nil {
			return
		}
	} else {
		if err = DB.Socket.Where("hh_id=?", scheduleID).First(&schdules).Error; err != nil {
			return
		}
	}
	return
}

func GetSchedulesList() (schedules Schedules, err error) {
	if err = DB.Socket.Find(&schedules).Error; err != nil {
		err = fmt.Errorf("bd GetSchedulesList getting Error: %w", err)
	}
	return
}

// Pool of vacancie search keys from DB getting
func GetVacancyPatterns() (vacNames VacancyNamePatterns, err error) {
	if err = DB.Socket.Find(&vacNames).Error; err != nil {
		err = fmt.Errorf("vacancie name poll getting error: %w", err)
		return nil, err
	}
	return vacNames, nil
}

// -------------------------------------------------------<<<JobData-----------------------
func (ja JobAnnounces) SaveInDB() (err error) {
	if err = DB.Socket.Save(&ja).Error; err != nil {
		err = fmt.Errorf("job announces update error: %w", err)
	}
	return
}

func (ud UserData) GetJobAnnounces(areas Countries) (announces JobAnnounces, err error) {
	var expierence string
	if ud.ExperienceYear < 1 {
		expierence = "noExperience"
	} else if ud.ExperienceYear >= 1 && ud.ExperienceYear <= 3 {
		expierence = "between1And3"
	} else if ud.ExperienceYear < 3 && ud.ExperienceYear <= 6 {
		expierence = "between3And6"
	} else if ud.ExperienceYear > 6 {
		expierence = "moreThan6"
	}

	var shownAnnounces []UserPivotVacancy
	if err = DB.Socket.Where("uid=?", ud.TgID).Find(&shownAnnounces).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("shown announces getting error: %w", err)
			return
		}

	}

	var shownAnnouncesIDs []uint
	for _, shown := range shownAnnounces {
		shownAnnouncesIDs = append(shownAnnouncesIDs, shown.JobID)
	}

	locationsTarget := areas.FindContainLocationIDsList(ud.Location)
	if len(locationsTarget) == 0 {
		if len(shownAnnounces) != 0 {
			if err = DB.Socket.Limit(50).Where("LOWER(name) like ? and expierence = ? and schedule = ? and  item_id not in ? ", "%"+strings.ToLower(ud.VacancyName)+"%", expierence, ud.Schedule, shownAnnouncesIDs).Find(&announces).Error; err != nil {
				err = fmt.Errorf("db vacancy with param schedule getting error: %w", err)
				return
			}

		} else {

			if err = DB.Socket.Limit(50).Where("LOWER(name) like ? and expierence = ? and schedule = ?", "%"+strings.ToLower(ud.VacancyName)+"%", expierence, ud.Schedule).Find(&announces).Error; err != nil {
				err = fmt.Errorf("db vacancy with param schedule getting error: %w", err)
				return
			}

		}
	} else {
		if len(shownAnnounces) != 0 {

			if err = DB.Socket.Limit(50).Where("LOWER(name) like ? and expierence = ? and schedule = ? and  item_id not in ? and area in ?", "%"+strings.ToLower(ud.VacancyName)+"%", expierence, ud.Schedule, shownAnnouncesIDs, locationsTarget).Find(&announces).Error; err != nil {
				err = fmt.Errorf("db vacancy with param schedule getting error: %w", err)
				return
			}

		} else {

			if err = DB.Socket.Limit(50).Where("LOWER(name) like ? and expierence = ? and schedule = ?  and area in ?", "%"+strings.ToLower(ud.VacancyName)+"%", expierence, ud.Schedule, locationsTarget).Find(&announces).Error; err != nil {
				err = fmt.Errorf("db vacancy with param schedule getting error: %w", err)
				return
			}

		}
	}

	return
}

// ------------------------------------------------------->>>JobData-----------------------

func CreatePivotVacancyAnnouncesAndUserIds(jobAnnouncesIDs []uint, uid uint) (err error) {
	var tempPivot []UserPivotVacancy
	for _, id := range jobAnnouncesIDs {
		tempPivot = append(tempPivot, UserPivotVacancy{UID: uid, JobID: id})
	}
	if err = DB.Socket.Create(&tempPivot).Error; err != nil {
		err = fmt.Errorf("db vacancy-user pivot record writing error: %w", err)
	}
	return
}
