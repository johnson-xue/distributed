package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
}

// 注册
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	//注册成功后请求此注册服务需要的已注册服务
	err := r.sendRequiredServices(reg)
	r.notify(patch{
		Added: []patchEntry{
			{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})
	return err
}

func (r registry) notify(fullPatch patch) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	//对已经注册的服务进行循环，将服务传入协程里面
	for _, reg := range r.registrations {
		go func(reg Registration) {
			//对已注册服务所依赖的服务进行循环
			for _, reqService := range reg.RequiredServices {
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				//if true 有需要更新的地方
				sendUpdate := false
				//更新分类：添加了那些服务
				for _, added := range fullPatch.Added {
					//添加的服务正好是所依赖的服务，则添加到p的切片里面
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				//更新分类：移除了那些服务
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				//如果有更新发送URl
				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						log.Panicln(err)
						return
					}
				}
			}
		}(reg)
	}
}

func (r *registry) sendRequiredServices(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	//遍历已经注册的服务
	for _, serviceReg := range r.registrations {
		//遍历请求的服务
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
	for i := range reg.registrations {
		if reg.registrations[i].ServiceURL == url {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: r.registrations[i].ServiceName,
						URL:  r.registrations[i].ServiceURL,
					},
				},
			})
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], r.registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

var reg = registry{
	registrations: make([]Registration, 0),
	mutex:         new(sync.RWMutex),
}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter,
	r *http.Request) {
	log.Println("Request received")
	switch r.Method {
	//post请求
	case http.MethodPost:
		//解码器
		dec := json.NewDecoder(r.Body)
		var registration Registration
		//解码
		err := dec.Decode(&registration)
		//出现错误
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v with URL: %v\n", registration.ServiceName,
			registration.ServiceURL)
		err = reg.add(registration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL: %s", url)
		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	//不是post请求
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
