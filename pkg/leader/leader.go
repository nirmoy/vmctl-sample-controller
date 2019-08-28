package leader

import (
	"context"
	"time"

	"github.com/google/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
)

var (
	leaseLockName      = "simple-controller-lock"
	leaseLockNamespace = "kube-system"
)

type Leader struct {
	kubeClientSet kubernetes.Interface
	uuid          string
	isLeader      bool
	cancel        context.CancelFunc
	lock          *resourcelock.LeaseLock
}

func LeaderInit(kubeclientset kubernetes.Interface) *Leader {
	uuid := uuid.New().String()
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseLockName,
			Namespace: leaseLockNamespace,
		},
		Client: kubeclientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: uuid,
		},
	}
	return &Leader{
		kubeClientSet: kubeclientset,
		uuid:          uuid,
		lock:          lock,
		isLeader:      false,
	}
}
func (l *Leader) StartElection(notify chan int) {
	ctx, cancel := context.WithCancel(context.Background())
	// Need to keep host of this to release the lease on controller.Run()'s exit
	l.cancel = cancel

	go leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: l.lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   5 * time.Second,
		RenewDeadline:   2 * time.Second,
		RetryPeriod:     1 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				l.isLeader = true
				klog.Infof("%s: leading", l.uuid)
				notify <- 1
			},
			OnStoppedLeading: func() {
				l.isLeader = false
				klog.Infof("%s: lost lease", l.uuid)
				notify <- 1
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == l.uuid {
					l.isLeader = true
					notify <- 1
				}
				klog.Infof("new leader elected: %v", identity)
			},
		},
	})
	return
}

func (l *Leader) IsLeader() bool {
	return l.isLeader
}

func (l *Leader) Clean() {
	l.cancel()
}

/*
var (
	configMapName   = "simple-controller-map"
	configMapSchema = &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      configMapName,
			Namespace: "default",
			Annotations: map[string]string{
				"lock-annotation": "",
			},
		},
	}
)

type Lock struct {
	kubeClientSet kubernetes.Interface
	uuid          string
}

func (l *Lock) LockInit(kubeclientset kubernetes.Interface) error {
	l.kubeClientSet = kubeclientset
	l.uuid = uuid.New().String()
	_, err := l.kubeClientSet.CoreV1().ConfigMaps("default").Get(configMapName, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		return nil
	}
	//configMapSchema.Annotations["lock-annotation"] = l.uuid
	_, err = l.kubeClientSet.CoreV1().ConfigMaps("default").Create(configMapSchema)
	if err != nil {
		return err
	}

	return nil
}

func (l *Lock) Lock() error {
	configMap, err := l.kubeClientSet.CoreV1().ConfigMaps("default").Get(configMapName, v1.GetOptions{})
	if err != nil {
		return err
	}
	if configMap.Annotations["lock-annotation"] == "" {
		oldConfig, err := json.Marshal(configMap)
		if err != nil {
			fmt.Println(err)
			return err
		}
		configMap.Annotations["lock-annotation"] = l.uuid
		newConfig, err := json.Marshal(configMap)
		if err != nil {
			fmt.Println(err)
			return err
		}
		patch, err := strategicpatch.CreateTwoWayMergePatch(oldConfig, newConfig, corev1.ConfigMap{})
		if err != nil {
			fmt.Printf("Fail to create patch: %v", err)
			return err
		}
		_, err = l.kubeClientSet.CoreV1().ConfigMaps("default").Patch(configMap.Name, types.StrategicMergePatchType, patch)
		if err != nil {
			fmt.Printf("Fail to patch : %v", err)
			return err
		}
		return nil
	}
	return fmt.Errorf("already locked ")
}

func (l *Lock) Unlock() error {
	configMap, err := l.kubeClientSet.CoreV1().ConfigMaps("default").Get(configMapName, v1.GetOptions{})
	if err != nil {
		return err
	}
	if configMap.Annotations["lock-annotation"] == l.uuid {
		oldConfig, err := json.Marshal(configMap)
		if err != nil {
			fmt.Println(err)
			return err
		}
		configMap.Annotations["lock-annotation"] = ""
		newConfig, err := json.Marshal(configMap)
		if err != nil {
			fmt.Println(err)
			return err
		}
		patch, err := strategicpatch.CreateTwoWayMergePatch(oldConfig, newConfig, corev1.ConfigMap{})
		if err != nil {
			fmt.Printf("Fail to create patch: %v", err)
			return err
		}
		_, err = l.kubeClientSet.CoreV1().ConfigMaps("default").Patch(configMap.Name, types.StrategicMergePatchType, patch)
		if err != nil {
			fmt.Printf("Fail to patch : %v", err)
			return err
		}
		return nil
	}
	return fmt.Errorf("Locked by someone else")
}
*/
