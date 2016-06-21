/*
 * MumbleDJ
 * By Matthieu Grieger
 * commands/addnext.go
 * Copyright (c) 2016 Matthieu Grieger (MIT License)
 */

package commands

import (
	"errors"
	"fmt"

	"github.com/layeh/gumble/gumble"
	"github.com/matthieugrieger/mumbledj/interfaces"
	"github.com/spf13/viper"
)

// AddNextCommand is a command that adds an audio track associated with a supported
// URL to the queue as the next item.
type AddNextCommand struct{}

// Aliases returns the current aliases for the command.
func (c *AddNextCommand) Aliases() []string {
	return viper.GetStringSlice("commands.addnext.aliases")
}

// Description returns the description for the command.
func (c *AddNextCommand) Description() string {
	return viper.GetString("commands.addnext.description")
}

// IsAdminCommand returns true if the command is only for admin use, and
// returns false otherwise.
func (c *AddNextCommand) IsAdminCommand() bool {
	return viper.GetBool("commands.addnext.is_admin")
}

// Execute executes the command with the given user and arguments.
// Return value descriptions:
//    string: A message to be returned to the user upon successful execution.
//    bool:   Whether the message should be private or not. true = private,
//            false = public (sent to whole channel).
//    error:  An error message to be returned upon unsuccessful execution.
//            If no error has occurred, pass nil instead.
// Example return statement:
//    return "This is a private message!", true, nil
func (c *AddNextCommand) Execute(user *gumble.User, args ...string) (string, bool, error) {
	var (
		allTracks      []interfaces.Track
		tracks         []interfaces.Track
		service        interfaces.Service
		err            error
		lastTrackAdded interfaces.Track
	)

	if len(args) == 0 {
		return "", true, errors.New("A URL must be supplied with the addnext command")
	}

	for _, arg := range args {
		if service, err = DJ.GetService(arg); err == nil {
			tracks, err = service.GetTracks(arg, user)
			if err == nil {
				allTracks = append(allTracks, tracks...)
			}
		}
	}

	if len(allTracks) == 0 {
		return "", true, errors.New("No valid tracks were found with the provided URL(s)")
	}

	numTooLong := 0
	numAdded := 0
	// We must loop backwards here to preserve the track order when inserting tracks.
	for i := len(allTracks) - 1; i >= 0; i-- {
		if err = DJ.Queue.InsertTrack(1, allTracks[i]); err != nil {
			numTooLong++
		} else {
			numAdded++
			lastTrackAdded = allTracks[i]
		}
	}

	if numAdded == 0 {
		return "", true, errors.New("Your track(s) were either too long or an error occurred while processing them. No track(s) have been added.")
	} else if numAdded == 1 {
		return fmt.Sprintf("<b>%s</b> added <b>1</b> track to the queue:<br>\"%s\" from %s",
			user.Name, lastTrackAdded.GetTitle(), lastTrackAdded.GetService()), false, nil
	}

	retString := fmt.Sprintf("<b>%s</b> added <b>%d</b> tracks to the queue.", user.Name, numAdded)
	if numTooLong != 0 {
		retString += fmt.Sprintf("<br><b>%d</b> tracks could not be added due to error or because they are too long.", numTooLong)
	}
	return retString, false, nil
}