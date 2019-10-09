// Manage the watcher in elasticsearch
// API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html
// Supported version:
//  - v6
//  - v7
package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	elastic7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Watcher Json object
type Watcher struct {
	Watcher *WatcherSpec `json:"watch"`
}
type WatcherSpec struct {
	Trigger        interface{} `json:"trigger,omitempty"`
	Input          interface{} `json:"input,omitempty"`
	Condition      interface{} `json:"condition,omitempty"`
	Actions        interface{} `json:"actions,omitempty"`
	Metadata       interface{} `json:"metadata,omitempty"`
	ThrottlePeriod string      `json:"throttle_period,omitempty"`
}

// Resource specification to handle watcher in Elasticsearch
func resourceElasticsearchWatcher() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchWatcherCreate,
		Read:   resourceElasticsearchWatcherRead,
		Update: resourceElasticsearchWatcherUpdate,
		Delete: resourceElasticsearchWatcherDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"trigger": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"input": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"condition": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"actions": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
			"throttle_period": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJson,
			},
		},
	}
}

// Create new watcher in Elasticsearch
func resourceElasticsearchWatcherCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	err := createWatcher(d, meta)
	if err != nil {
		return err
	}
	d.SetId(name)

	log.Infof("Created watcher %s successfully", name)

	return resourceElasticsearchWatcherRead(d, meta)
}

// Read existing watch in Elasticsearch
func resourceElasticsearchWatcherRead(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	var b []byte

	log.Debugf("Watcher id:  %s", id)

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// v6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.WatcherGetWatch(
			id,
			client.API.XPack.WatcherGetWatch.WithContext(context.Background()),
			client.API.XPack.WatcherGetWatch.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Watcher %s not found - removing from state", id)
				log.Warnf("Watcher %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when get watcher %s: %s", id, res.String())
			}
		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

	// v7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Watcher.GetWatch(
			id,
			client.API.Watcher.GetWatch.WithContext(context.Background()),
			client.API.Watcher.GetWatch.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Watcher %s not found - removing from state", id)
				log.Warnf("Watcher %s not found - removing from state", id)
				d.SetId("")
				return nil
			} else {
				return errors.Errorf("Error when get watcher %s: %s", id, res.String())
			}
		}
		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	default:
		return errors.New("Watcher is only supported by the elastic library >= v6!")
	}

	log.Debugf("Get watcher %s successfully:\n%s", id, string(b))
	watcher := &Watcher{}
	err := json.Unmarshal(b, watcher)
	if err != nil {
		return err
	}

	watcherSpec := watcher.Watcher

	log.Debugf("Watcher %+v", watcherSpec)

	d.Set("name", id)
	d.Set("trigger", watcherSpec.Trigger)
	d.Set("input", watcherSpec.Trigger)
	d.Set("condition", watcherSpec.Condition)
	d.Set("actions", watcherSpec.Actions)
	d.Set("metadata", watcherSpec.Metadata)
	d.Set("throttle_period", watcherSpec.ThrottlePeriod)

	log.Infof("Read watcher %s successfully", id)

	return nil
}

// Update existing watcher in Elasticsearch
func resourceElasticsearchWatcherUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createWatcher(d, meta)
	if err != nil {
		return err
	}

	log.Infof("Updated watcher %s successfully", d.Id())

	return resourceElasticsearchWatcherRead(d, meta)
}

// Delete existing role in Elasticsearch
func resourceElasticsearchWatcherDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()
	log.Debugf("Watcher id: %s", id)

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// v6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.WatcherDeleteWatch(
			id,
			client.API.XPack.WatcherDeleteWatch.WithContext(context.Background()),
			client.API.XPack.WatcherDeleteWatch.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Watcher %s not found - removing from state", id)
				log.Warnf("Watcher %s not found - removing from state", id)
				d.SetId("")
				return nil

			}
			return errors.Errorf("Error when delete watcher %s: %s", id, res.String())
		}

	// v7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Watcher.DeleteWatch(
			id,
			client.API.Watcher.DeleteWatch.WithContext(context.Background()),
			client.API.Watcher.DeleteWatch.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			if res.StatusCode == 404 {
				fmt.Printf("[WARN] Watcher %s not found - removing from state", id)
				log.Warnf("Watcher %s not found - removing from state", id)
				d.SetId("")
				return nil

			}
			return errors.Errorf("Error when delete watcher %s: %s", id, res.String())
		}

	default:
		return errors.New("Watcher is only supported by the elastic library >= v6!")
	}

	d.SetId("")

	log.Infof("Deleted watcher %s successfully", id)
	return nil

}

// Print Watcher object as Json string
func (r *WatcherSpec) String() string {
	json, _ := json.Marshal(r)
	return string(json)
}

// Create or update watcher in Elasticsearch
func createWatcher(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	trigger := optionalInterfaceJson(d.Get("trigger").(string))
	input := optionalInterfaceJson(d.Get("input").(string))
	condition := optionalInterfaceJson(d.Get("condition").(string))
	actions := optionalInterfaceJson(d.Get("actions").(string))
	metadata := optionalInterfaceJson(d.Get("metadata").(string))
	throttlePeriod := d.Get("throttle_period").(string)

	watcher := &WatcherSpec{
		Trigger:        trigger,
		Input:          input,
		Condition:      condition,
		Actions:        actions,
		Metadata:       metadata,
		ThrottlePeriod: throttlePeriod,
	}
	log.Debug("Name: ", name)
	log.Debug("Watcher: ", watcher)

	data, err := json.Marshal(watcher)
	if err != nil {
		return err
	}

	// Use the right client depend to Elasticsearch version
	switch meta.(type) {
	// v6
	case *elastic6.Client:
		client := meta.(*elastic6.Client)
		res, err := client.API.XPack.WatcherPutWatch(
			name,
			client.API.XPack.WatcherPutWatch.WithBody(bytes.NewReader(data)),
			client.API.XPack.WatcherPutWatch.WithContext(context.Background()),
			client.API.XPack.WatcherPutWatch.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add watcher %s: %s", name, res.String())
		}

	// v7
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.Watcher.PutWatch(
			name,
			client.API.Watcher.PutWatch.WithBody(bytes.NewReader(data)),
			client.API.Watcher.PutWatch.WithContext(context.Background()),
			client.API.Watcher.PutWatch.WithPretty(),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			return errors.Errorf("Error when add watcher %s: %s", name, res.String())
		}
	default:
		return errors.New("Watcher is only supported by the elastic library >= v6!")
	}

	return nil
}
