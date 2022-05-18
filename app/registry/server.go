package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	ServerPort  = ":3000"
	ServicesURL = "http://registry" + ServerPort + "/services"
)

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
}

func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()

	err := r.sendRequiredServices(reg)

	r.notify(patch{
		Added: []patchEntry{
			patchEntry{Name: reg.ServiceName,
				URL: reg.ServiceURL},
		},
	})

	return err
}

func (r registry) notify(fullPatch patch) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, reg := range r.registrations {
		go func(reg Registration) {
			for _, reqService := range reg.RequiredServices {
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				sendUpdate := false
				for _, added := range fullPatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
					}
					sendUpdate = true
				}

				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
					}
					sendUpdate = true
				}

				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
}

func (r registry) sendRequiredServices(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil

}

func (r registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}

	return nil

}

func (r *registry) remove(url string) error {
	defer func() {
		fmt.Println("currently have registrations ")
		fmt.Println(r.registrations)
	}()
	fmt.Println(r.registrations)
	for i := 0; i < len(r.registrations); i++ {
		fmt.Println("comparison")
		fmt.Println(r.registrations[i].ServiceURL, " ", url)
		if r.registrations[i].ServiceURL == url {
			fullPatch := patch{Removed: []patchEntry{
				patchEntry{Name: r.registrations[i].ServiceName,
					URL: r.registrations[i].ServiceURL,
				},
			},
			}
			fmt.Printf("%+v\n", fullPatch)
			r.notify(fullPatch)
			r.mutex.Lock()
			r.registrations[i] = r.registrations[len(r.registrations)-1]
			r.registrations = r.registrations[:len(r.registrations)-1]
			r.mutex.Unlock()
			return nil
		}
	}

	return fmt.Errorf("Service at URL %v not found\n", url)

}

var reg = registry{registrations: make([]Registration, 0),
	mutex: new(sync.RWMutex)}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")

	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var regis Registration
		err := dec.Decode(&regis)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		log.Printf("Adding service: %v with URL: %v \n", regis.ServiceName, regis.ServiceURL)

		err = reg.add(regis)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url := string(payload)
		log.Printf("Removing service at url %v\n", url)
		//err = reg.remove(url)
		//if err != nil {
		//	log.Println(err)
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			go func(reg Registration) {
				defer wg.Done()
				success := true
				fmt.Println("CALLING ", reg.HeartbeatURL)
				for attempts := 0; attempts < 3; attempts++ {
					res, err := http.Get(reg.HeartbeatURL)
					if err != nil {
						log.Println(err)
					} else if res.StatusCode == http.StatusOK {
						log.Printf("Heartbeat check passed for %v\n", reg.ServiceName)
						if !success {
							r.add(reg)
						}
						break

					}
					log.Printf("Heartbeat check failed for %v\n", reg.ServiceName)
					if success {
						success = false
						r.remove(reg.ServiceURL)
					}
					time.Sleep(2 * time.Second)
				}
			}(reg)
		}
		wg.Wait()
		time.Sleep(freq)

	}
}

var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartbeat(3 * time.Second)
	})
}