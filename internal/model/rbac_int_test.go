package model

// BOZO!!
// func TestParseRules(t *testing.T) {
// 	ok, nok := toVerbIcon(true), toVerbIcon(false)
// 	_ = nok

// 	uu := []struct {
// 		pp []rbacv1.PolicyRule
// 		e  render.Rows
// 	}{
// 		{
// 			[]rbacv1.PolicyRule{
// 				{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}},
// 			},
// 			render.Rows{
// 				render.Row{Fields: render.Fields{"*.*", "*", ok, ok, ok, ok, ok, ok, ok, ok, ""}},
// 			},
// 		},
// 		{
// 			[]rbacv1.PolicyRule{
// 				{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"get"}},
// 			},
// 			render.Rows{
// 				render.Row{Fields: render.Fields{"*.*", "*", ok, nok, nok, nok, nok, nok, nok, nok, ""}},
// 			},
// 		},
// 		{
// 			[]rbacv1.PolicyRule{
// 				{APIGroups: []string{""}, Resources: []string{"*"}, Verbs: []string{"list"}},
// 			},
// 			render.Rows{
// 				render.Row{Fields: render.Fields{"*", "v1", nok, ok, nok, nok, nok, nok, nok, nok, ""}},
// 			},
// 		},
// 		{
// 			[]rbacv1.PolicyRule{
// 				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"list"}, ResourceNames: []string{"fred"}},
// 			},
// 			render.Rows{
// 				render.Row{Fields: render.Fields{"pods", "v1", nok, ok, nok, nok, nok, nok, nok, nok, ""}},
// 				render.Row{Fields: render.Fields{"pods/fred", "v1", nok, ok, nok, nok, nok, nok, nok, nok, ""}},
// 			},
// 		},
// 		{
// 			[]rbacv1.PolicyRule{
// 				{APIGroups: []string{}, Resources: []string{}, Verbs: []string{"get"}, NonResourceURLs: []string{"/fred"}},
// 			},
// 			render.Rows{
// 				render.Row{Fields: render.Fields{"/fred", resource.NAValue, ok, nok, nok, nok, nok, nok, nok, nok, ""}},
// 			},
// 		},
// 		{
// 			[]rbacv1.PolicyRule{
// 				{APIGroups: []string{}, Resources: []string{}, Verbs: []string{"get"}, NonResourceURLs: []string{"fred"}},
// 			},
// 			render.Rows{
// 				render.Row{Fields: render.Fields{"/fred", resource.NAValue, ok, nok, nok, nok, nok, nok, nok, nok, ""}},
// 			},
// 		},
// 	}

// 	var v Rbac
// 	for _, u := range uu {
// 		evts := v.parseRules(u.pp)
// 		for k, v := range u.e {
// 			assert.Equal(t, v, evts[k].Fields)
// 		}
// 	}
// }
