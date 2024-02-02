package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"path"
)

type PathItemsStruct struct {
	Apps string `json:"apps"`
}

type ConfigStruct struct {
	Path        PathItemsStruct `json:"path"`
	DryRun      bool            `json:"dryrun"`
	RecoverMode bool            `json:"recmode"`
}

func main() {
	var pConfigPath = flag.String("config", "./config.json", "The path of the config.")
	flag.Parse()

	log.Printf("config: %v", *pConfigPath)

	var data []byte
	var err error
	if data, err = os.ReadFile(*pConfigPath); err != nil {
		log.Fatal("[FATAL] Could not found config.json")
	}

	var config *ConfigStruct
	if err = json.Unmarshal(data, &config); err != nil {
		log.Fatal("[FATAL] Could not unmarshal json content.")
	}

	var dryrun = config.DryRun
	var dir []os.DirEntry
	if dir, err = os.ReadDir(config.Path.Apps); err != nil {
		log.Fatalf("[FATAL] Failed to readdir: %v", config.Path.Apps)
	}

	// Walk folders in scoops/apps
	var rec = make(map[string]string)
	for _, e := range dir {
		if e.IsDir() {
			var installPath = path.Join(config.Path.Apps, e.Name(), "current", "install.json")
			if _, err = os.Stat(installPath); errors.Is(err, os.ErrNotExist) {
				log.Printf("[WARN] install.json not found, skipped: %v", e.Name())
				log.Print(err)
				continue
			}

			// Found install.json
			var installObject map[string]any
			var installData []byte

			if installData, err = os.ReadFile(installPath); err != nil {
				log.Printf("[WARN] Failed to read install.json in %v, skipped.", e.Name())
				log.Print(err)
				continue
			}

			if err = json.Unmarshal(installData, &installObject); err != nil {
				log.Printf("[WARN] Failed to unmarshal install.json in %v, skipped.", e.Name())
				log.Print(err)
				continue
			}

			var bucket = installObject["bucket"].(string)
			if bucket == "scoop-cn" {
				log.Printf("[INFO] The bucket of %v already set to scoop-cn, skipped", e.Name())
				continue
			}

			rec[e.Name()] = bucket
			installObject["bucket"] = "scoop-cn"

			if installData, err = json.Marshal(&installObject); err != nil {
				log.Printf("[WARN] Failed to rewrite install.json (json-marshal-failed) for %v", e.Name())
				log.Print(err)
				continue
			}

			if !dryrun {
				if err = os.WriteFile(installPath, installData, 0777); err != nil {
					log.Printf("[WARN] Failed to rewrite install.json for %v.", e.Name())
					log.Print(err)
				}
			} else {
				log.Print("[INFO] DRYRUN OUTPUT:")
				log.Print(string(installData))
			}
		}
	}

	// Process done
	log.Print("[INFO] All processes are done.")
	var recData []byte
	if recData, err = json.Marshal(&rec); err != nil {
		log.Printf("[ERROR] Could not write rec.json.")
		log.Print(err)
	}

	if !dryrun {
		if err = os.WriteFile("./rec.json", recData, 0777); err != nil {
			log.Printf("[ERROR] Could not write rec.json.")
			log.Print(err)
		}
	}

	log.Print(string(recData))
}
