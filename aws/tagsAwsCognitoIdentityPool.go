package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"regexp"
)

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsCognitoIdentityPool(conn *cognitoidentity.CognitoIdentity, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsCognitoIdentityPool(tagsFromMapCognitoIdentityPool(o), tagsFromMapCognitoIdentityPool(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			k := make([]*string, 0, len(remove))
			for i := range remove {
				k = append(k, &i)
			}

			_, err := conn.UntagResource(&cognitoidentity.UntagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				TagKeys:     k,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&cognitoidentity.TagResourceInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
				Tags:        create,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsCognitoIdentityPool(oldTags, newTags map[string]*string) (map[string]*string, map[string]*string) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for key, value := range newTags {
		create[key] = aws.StringValue(value)
	}

	// Build the list of what to remove
	var remove = make(map[string]*string)
	for k, v := range oldTags {
		old, ok := create[k]
		if !ok || old != aws.StringValue(v) {
			// Delete it!
			remove[k] = v
		} else if ok {
			delete(create, k)
		}
	}

	return tagsFromMapCognitoIdentityPool(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapCognitoIdentityPool(m map[string]interface{}) map[string]*string {
	result := make(map[string]*string)
	for key, value := range m {
		if !tagIgnoredCognitoIdentityPool(key, value.(string)) {
			result[key] = aws.String(value.(string))
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapCognitoIdentityPool(ts map[string]*string) map[string]string {
	result := make(map[string]string)
	for key, value := range ts {
		if !tagIgnoredCognitoIdentityPool(key, *value) {
			result[key] = aws.StringValue(value)
		}
	}

	return result
}

// compare a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnoredCognitoIdentityPool(tagKey string, tagValue string) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, tagKey)
		r, _ := regexp.MatchString(v, tagKey)
		if r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", tagKey, tagValue)
			return true
		}
	}
	return false
}

// getTags is a helper to get the tags for a resource. It expects the
// tags field to be named "tags"
func getTagsCognitoIdentityPool(conn *cognitoidentity.CognitoIdentity, d *schema.ResourceData, arn string) error {
	resp, err := conn.ListTagsForResource(&cognitoidentity.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	})

	if err != nil {
		return err
	}

	return d.Set("tags", tagsToMapCognitoIdentityPool(resp.Tags))
}
