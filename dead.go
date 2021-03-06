package main

import (
	"strings"
	"time"
)

type DeadState struct {
	CommandSuite
	start int64
}

func NewDeadState(died int64) ConnectionState {
	return &DeadState{
		start:        died,
		CommandSuite: CommandSet{},
	}
}

func (d *DeadState) Enter(c *Connection) {
	msg := `
Y88b   d88P                             d8888                                   
 Y88b d88P                             d88888                                   
  Y88o88P                             d88P888                                   
   Y888P  .d88b.  888  888           d88P 888 888d888 .d88b.                    
    888  d88""88b 888  888          d88P  888 888P"  d8P  Y8b                   
    888  888  888 888  888         d88P   888 888    88888888                   
    888  Y88..88P Y88b 888        d8888888888 888    Y8b.                       
    888   "Y88P"   "Y88888       d88P     888 888     "Y8888                    


                                     ____
                              __,---'     '--.__
                           ,-'                ; '.
                          ,'                  '--.'--.
                         ,'                       '._ '-.
                         ;                     ;     '-- ;
                       ,-'-_       _,-~~-.      ,--      '.
                       ;;   '-,;    ,'~'.__    ,;;;    ;  ;
                       ;;    ;,'  ,;;      ',  ;;;     '. ;
                       ':   ,'    ':;     __/  '.;      ; ;
                        ;~~^.   '.   '---'~~    ;;      ; ;
                        ',' '.   '.            .;;;     ;'
                        ,',^. '.  '._    __    ':;     ,'
                        '-' '--'    ~'--'~~'--.  ~    ,'
                       /;'-;_ ; ;. /. /   ; ~~'-.     ;
-._                   ; ;  ; ',;'-;__;---;      '----'
   '--.__             ''-'-;__;:  ;  ;__;
 ...     '--.__                '-- '-'
'--.:::...     '--.__                ____
    '--:::::--.      '--.__    __,--'    '.
        '--:::';....       '--'       ___  '.
            '--'-:::...     __           )  ;
                  ~'-:::...   '---.      ( ,'
                      ~'-:::::::::'--.   '-.
                          ~'-::::::::'.    ;
                              ~'--:::,'   ,'
                                   ~~'--'~

    8888888b.  8888888888        d8888 8888888b.  
    888  "Y88b 888              d88888 888  "Y88b 
    888    888 888             d88P888 888    888 
    888    888 8888888        d88P 888 888    888 
    888    888 888           d88P  888 888    888 
    888    888 888          d88P   888 888    888 
    888  .d88P 888         d8888888888 888  .d88P 
    8888888P"  8888888888 d88P     888 8888888P"  
`
	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		c.Write([]byte(line + "\n"))
		time.Sleep(20 * time.Millisecond)
	}
}

func (d *DeadState) Tick(c *Connection, frame int64) ConnectionState {
	if frame-d.start > options.respawnFrames {
		return c.game.SpawnPlayer()
	}
	return d
}

func (d *DeadState) Exit(c *Connection) {
	c.Printf("You're alive again.\n")
}

func (d *DeadState) String() string { return "dead" }

func (d *DeadState) FillStatus(c *Connection, s *status) {
	s.Description = `
Y88b   d88P                             d8888                                   
 Y88b d88P                             d88888                                   
  Y88o88P                             d88P888                                   
   Y888P  .d88b.  888  888           d88P 888 888d888 .d88b.                    
    888  d88""88b 888  888          d88P  888 888P"  d8P  Y8b                   
    888  888  888 888  888         d88P   888 888    88888888                   
    888  Y88..88P Y88b 888        d8888888888 888    Y8b.                       
    888   "Y88P"   "Y88888       d88P     888 888     "Y8888                    


                                     ____
                              __,---'     '--.__
                           ,-'                ; '.
                          ,'                  '--.'--.
                         ,'                       '._ '-.
                         ;                     ;     '-- ;
                       ,-'-_       _,-~~-.      ,--      '.
                       ;;   '-,;    ,'~'.__    ,;;;    ;  ;
                       ;;    ;,'  ,;;      ',  ;;;     '. ;
                       ':   ,'    ':;     __/  '.;      ; ;
                        ;~~^.   '.   '---'~~    ;;      ; ;
                        ',' '.   '.            .;;;     ;'
                        ,',^. '.  '._    __    ':;     ,'
                        '-' '--'    ~'--'~~'--.  ~    ,'
                       /;'-;_ ; ;. /. /   ; ~~'-.     ;
-._                   ; ;  ; ',;'-;__;---;      '----'
   '--.__             ''-'-;__;:  ;  ;__;
 ...     '--.__                '-- '-'
'--.:::...     '--.__                ____
    '--:::::--.      '--.__    __,--'    '.
        '--:::';....       '--'       ___  '.
            '--'-:::...     __           )  ;
                  ~'-:::...   '---.      ( ,'
                      ~'-:::::::::'--.   '-.
                          ~'-::::::::'.    ;
                              ~'--:::,'   ,'
                                   ~~'--'~

    8888888b.  8888888888        d8888 8888888b.  
    888  "Y88b 888              d88888 888  "Y88b 
    888    888 888             d88P888 888    888 
    888    888 8888888        d88P 888 888    888 
    888    888 888           d88P  888 888    888 
    888    888 888          d88P   888 888    888 
    888  .d88P 888         d8888888888 888  .d88P 
    8888888P"  8888888888 d88P     888 8888888P"  
`
}
