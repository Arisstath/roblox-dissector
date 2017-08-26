package main
import "github.com/gskartwii/rbxfile"
import "sync"

type InstanceList struct {
    CommonMutex *sync.Mutex
    EAddReferent *sync.Cond
    Instances map[string]*rbxfile.Instance
    AddCallbacks map[string][]func(*rbxfile.Instance)
}

func (l *InstanceList) TryGetInstance(ref Referent) *rbxfile.Instance {
    instance := l.Instances[string(ref)]
    if instance == nil {
        instance = &rbxfile.Instance{Reference: string(ref)}
        l.AddInstance(ref, instance)
    }
    return instance
}

func (l *InstanceList) WaitForInstance(ref Referent) *rbxfile.Instance {
    instance := l.Instances[string(ref)]
    for instance == nil {
        println("Waiting on", string(ref))
        l.EAddReferent.Wait()
        instance = l.Instances[string(ref)]
        println("Check on", string(ref))
        if instance != nil {
            println("Wait successful", string(ref))
            return instance
        }
    }
    return instance
}
func (l *InstanceList) AddInstance(ref Referent, instance *rbxfile.Instance) {
    l.Instances[string(ref)] = instance
    for _, callback := range l.AddCallbacks[string(ref)] {
        callback(instance)
    }

    l.EAddReferent.Broadcast()
}

func (l *InstanceList) OnAddInstance(ref Referent, callback func(*rbxfile.Instance)) {
    instance := l.Instances[string(ref)]
    if instance == nil {
        thisPend := l.AddCallbacks[string(ref)]
        thisPend = append(thisPend, callback)
    } else {
        callback(instance)
    }
}
