package model

import (
	"fmt"
	"member2/contrib/helper"
)

func MemberClosureInsert(nodeID, targetID string) string {

	t := "SELECT ancestor, " + nodeID + ", lvl+1,'" + meta.Prefix + "' FROM tbl_members_tree WHERE descendant = " + targetID + " UNION SELECT " + nodeID + "," + nodeID + ",0"
	query := "INSERT INTO tbl_members_tree (ancestor, descendant,lvl,prefix) (" + t + ")"

	return query
}

func MemberClosureGetParent(uid string) ([]string, error) {

	uids := []string{}
	query := fmt.Sprintf("SELECT ancestor FROM tbl_member_tree WHERE descendant='%s' ORDER BY lvl ASC", uid)

	err := meta.MerchantDB.Select(&uids, query)
	if err != nil {
		body := fmt.Errorf("%s,[%s]", err.Error(), query)
		return uids, pushLog(body, helper.DBErr)
	}

	return uids, nil
}
