package main

import (
	"fmt"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[[:alpha:]][[:alnum:]-_]{0,19}$`)

func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

type Profile struct {
	id   int
	name string
}

func (p *Profile) Create() error {
	_, err := db.Exec(`
        insert into profiles
        (name)
        values
        (?)
    ;`, p.name)
	if err != nil {
		return fmt.Errorf("unable to create profile: %v", err)
	}
	return nil
}

func profilesTable() {
	stmnt := `create table if not exists profiles (
        id integer not null primary key autoincrement,
        name text unique
    );`
	if _, err := db.Exec(stmnt); err != nil {
		log_error("couldn't create profiles table: %v", err)
	}
}

func loadProfile(name string) (*Profile, error) {
	row := db.QueryRow(`select * from profiles where name = ?`, name)
	var p Profile
	if err := row.Scan(&p.id, &p.name); err != nil {
		return nil, fmt.Errorf("unable to fetch profile from database: %v", err)
	}
	return &p, nil
}
