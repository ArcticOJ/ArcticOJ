//go:build headless

package main

import (
	"context"
	"fmt"
	"github.com/ArcticOJ/blizzard/v0/core"
	"github.com/ArcticOJ/blizzard/v0/db"
	"github.com/ArcticOJ/blizzard/v0/db/schema/user"
	"github.com/ArcticOJ/blizzard/v0/logger/debug"
	"github.com/ArcticOJ/blizzard/v0/permission"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"strconv"
	"strings"
)

func createUser() *cobra.Command {
	c := &cobra.Command{
		Use:   "create_user <handle> <password> <email>",
		Short: "create a new user",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, e := core.HashConfig.HashEncoded([]byte(args[1]))
			if e != nil {
				return e
			}
			if e != nil {
				return fmt.Errorf("invalid role_id, role_id must be a valid 16-bit integer")
			}
			u := &user.User{
				Handle:   strings.ToLower(args[0]),
				Email:    strings.ToLower(args[2]),
				Password: string(r),
			}

			var roleId = -1

			if role, e := cmd.Flags().GetInt("role"); e == nil && role != -1 {
				roleId = role
			}
			if name, e := cmd.Flags().GetString("name"); e == nil && name != "" {
				u.DisplayName = name
			}
			//if org, e := cmd.Flags().GetString("org"); e == nil && org != "" {
			//	u.Organizations = org
			//}
			if e := db.Database.RunInTx(cmd.Context(), nil, func(ctx context.Context, tx bun.Tx) error {
				_, err := tx.NewInsert().Model(u).Returning("id").Exec(ctx)
				if err != nil {
					return err
				}
				if roleId >= 0 {
					_, err = tx.NewInsert().Model(&user.RoleMembership{
						RoleID: uint16(roleId),
						UserID: u.ID,
					}).Exec(ctx)
					if err != nil {
						return err
					}
				}
				return nil
			}); e != nil {
				return e
			}
			return nil
		},
	}
	c.Flags().String("name", "", "display name for created user")
	//c.Flags().String("org", "", "user's affiliated organization")
	c.Flags().Int("role", -1, "role to assign")
	return c
}

func createRole() *cobra.Command {
	c := &cobra.Command{
		Use:  "create_role <name> <priority> [permissions...]",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			perms := args[1:]
			var permBits permission.Permission = 0
			priority, e := strconv.Atoi(args[1])
			if e != nil {
				return e
			}
			for _, p := range perms {
				permBits |= permission.StringToPermission(strings.ToLower(p))
			}
			r := &user.Role{
				Name:        args[0],
				Priority:    uint16(priority),
				Permissions: permBits,
			}
			if icon, e := cmd.Flags().GetString("icon"); e == nil && icon != "" {
				r.Icon = icon
			}
			if color, e := cmd.Flags().GetString("color"); e == nil && color != "" {
				r.Color = color
			}
			_, err := db.Database.NewInsert().Model(r).Exec(cmd.Context())
			if err != nil {
				return err
			}
			debug.Dump()
			return nil
		},
	}
	c.Flags().String("icon", "", "icon of role")
	c.Flags().String("color", "", "color of the role")
	return c
}

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "manage database-related stuff",
}

func init() {
	manageCmd.AddCommand(
		createUser(),
		createRole(),
	)
}
