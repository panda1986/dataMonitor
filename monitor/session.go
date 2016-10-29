package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var globalSessions *Manager

func initSession() {
	globalSessions = NewManager("kl_monitor_key")
}

// global session manager
type Manager struct {
	cookieName string     //private cookiename
	lock       sync.Mutex // protects session
	provider   *MemoryProvider
}

func NewManager(cookieName string) *Manager {
	return &Manager{
		provider:   NewMemoryProvider(),
		cookieName: cookieName,
	}
}

// generate global session id by key of usercenter and username, timestamp, ip
func (manager *Manager) sessionId(uid interface{}, name, ip string) string {
	return Md5Hash(fmt.Sprintf("%v/%v/%v/%v", uid, name, ip, time.Now().Unix()))
}

// when login success, create session, set cookie
func (manager *Manager) SessionCreate(uid interface{}, name string, w http.ResponseWriter, r *http.Request) (sess *MemorySession) {
	// if cookie has content, update it
	// if not, create a session
	manager.lock.Lock()
	defer manager.lock.Unlock()

	sid := manager.sessionId(uid, name, r.RemoteAddr)
	sess = manager.provider.SessionInit(sid)
	sess.SetUser(name)
	sess.SetUserIp(r.RemoteAddr)
	cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true}
	http.SetCookie(w, &cookie)

	return sess
}

//Destroy sessionid
// when log out
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		manager.lock.Lock()
		defer manager.lock.Unlock()
		manager.provider.SessionDestroy(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

func (manager *Manager) SessionRead(w http.ResponseWriter, r *http.Request) (sess *MemorySession, err error) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return nil, fmt.Errorf("didn't have key")
	}

	sid := cookie.Value
	if sess, err = manager.provider.SessionRead(sid); err != nil {
		return nil, fmt.Errorf("read session failed, err is %v", err)
	}

	return
}

type MemoryProvider struct {
	sessions map[string]*MemorySession
	lock     *sync.Mutex
}

func NewMemoryProvider() *MemoryProvider {
	v := &MemoryProvider{
		sessions: make(map[string]*MemorySession),
		lock:     &sync.Mutex{},
	}
	return v
}

func (v *MemoryProvider) SessionInit(sid string) *MemorySession {
	v.lock.Lock()
	defer v.lock.Unlock()

	newSess := &MemorySession{sid: sid, timeAccessed: time.Now()}
	v.sessions[sid] = newSess

	return newSess
}

func (v *MemoryProvider) SessionRead(sid string) (*MemorySession, error) {
	if sess, ok := v.sessions[sid]; !ok {
		return nil, fmt.Errorf("read session, sid not exist in memory sessions")
	} else {
		return sess, nil
	}
}

func (v *MemoryProvider) SessionDestroy(sid string) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if _, ok := v.sessions[sid]; !ok {
		err := fmt.Errorf("delete session, sid %v not exist in memory sessions", sid)
		ol.E(nil, "session destory failed, err is %v", err)
		return err
	}

	delete(v.sessions, sid)

	return nil
}

type MemorySession struct {
	sid          string    //session id唯一标示
	timeAccessed time.Time //最后访问时间
	userName     string    //用户名
	ip           string    //用户ip
}

func (v *MemorySession) String() string {
	return fmt.Sprintf("sid:%v update:%v, user:%v ip:%v", v.sid, v.timeAccessed.Format("2006-01-02 15:04:05"), v.userName, v.ip)
}

// set value and update info.
func (v *MemorySession) SetUser(user string) {
	v.userName = user
}

func (v *MemorySession) SetUserIp(ip string) {
	v.ip = ip
}

func (v *MemorySession) GetUser() string {
	return v.userName
}

func (v *MemorySession) SessionID() string {
	return v.sid
}

func (v *MemorySession) Expired(to *time.Time) bool {
	return v.timeAccessed.Before(*to)
}

func Md5Hash(element string) string {
	h := md5.New()
	h.Write([]byte(element))

	return hex.EncodeToString(h.Sum(nil))
}
