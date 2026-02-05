package installer

import (
	"context"
	"errors"
	"testing"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/annotations"
	"github.com/redhat-appstudio/helmet/internal/k8s"

	o "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestNewJob tests the constructor
func TestNewJob(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := api.NewAppContext("test-app")
	kube := k8s.NewFakeKube()

	job := NewJob(appCtx, kube)

	g.Expect(job).ToNot(o.BeNil())
	g.Expect(job.appName).To(o.Equal("test-app"))
	g.Expect(job.retries).To(o.Equal(int32(0)))
}

// TestJob_LabelSelector tests the LabelSelector method
func TestJob_LabelSelector(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := api.NewAppContext("test-app")
	kube := k8s.NewFakeKube()
	job := NewJob(appCtx, kube)

	selector := job.LabelSelector()
	g.Expect(selector).To(o.Equal("installer-job." + annotations.RepoURI))
}

// TestJob_GetJobLogFollowCmd tests the GetJobLogFollowCmd method
func TestJob_GetJobLogFollowCmd(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := api.NewAppContext("test-app")
	kube := k8s.NewFakeKube()
	job := NewJob(appCtx, kube)

	cmd := job.GetJobLogFollowCmd("test-namespace")
	g.Expect(cmd).To(o.ContainSubstring("oc --namespace=test-namespace"))
	g.Expect(cmd).To(o.ContainSubstring("logs --follow"))
	g.Expect(cmd).To(o.ContainSubstring("type=" + job.LabelSelector()))
}

// TestJob_getJob tests the getJob method
func TestJob_getJob(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := api.NewAppContext("test-app")

	t.Run("no_job_found", func(t *testing.T) {
		kube := k8s.NewFakeKube()
		job := NewJob(appCtx, kube)

		result, err := job.getJob(context.Background())
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(errors.Is(err, ErrJobNotFound)).To(o.BeTrue())
		g.Expect(result).To(o.BeNil())
	})

	t.Run("single_job_found", func(t *testing.T) {
		testJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
		}

		kube := k8s.NewFakeKube(testJob)
		job := NewJob(appCtx, kube)

		result, err := job.getJob(context.Background())
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result.Name).To(o.Equal("test-job"))
	})

	t.Run("multiple_jobs_found", func(t *testing.T) {
		testJob1 := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job-1",
				Namespace: "namespace-1",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
		}
		testJob2 := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job-2",
				Namespace: "namespace-2",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
		}

		kube := k8s.NewFakeKube(testJob1, testJob2)
		job := NewJob(appCtx, kube)

		result, err := job.getJob(context.Background())
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("multiple installer jobs found"))
		g.Expect(result).To(o.BeNil())
	})
}

// TestJob_GetState tests the GetState method
func TestJob_GetState(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := api.NewAppContext("test-app")

	t.Run("state_not_found", func(t *testing.T) {
		kube := k8s.NewFakeKube()
		job := NewJob(appCtx, kube)

		_, err := job.GetState(context.Background())
		// ErrJobNotFound is returned by getJob, but GetState should handle it
		g.Expect(err).To(o.HaveOccurred())
	})

	t.Run("state_deploying", func(t *testing.T) {
		testJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
								Args:  []string{"deploy"},
							},
						},
					},
				},
			},
			Status: batchv1.JobStatus{
				Active: 1,
			},
		}

		kube := k8s.NewFakeKube(testJob)
		job := NewJob(appCtx, kube)

		state, err := job.GetState(context.Background())
		g.Expect(err).To(o.Succeed())
		g.Expect(state).To(o.Equal(Deploying))
	})

	t.Run("state_failed", func(t *testing.T) {
		testJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
								Args:  []string{"deploy"},
							},
						},
					},
				},
			},
			Status: batchv1.JobStatus{
				Failed: 1,
			},
		}

		kube := k8s.NewFakeKube(testJob)
		job := NewJob(appCtx, kube)

		state, err := job.GetState(context.Background())
		g.Expect(err).To(o.Succeed())
		g.Expect(state).To(o.Equal(Failed))
	})

	t.Run("state_done", func(t *testing.T) {
		testJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
								Args:  []string{"deploy"},
							},
						},
					},
				},
			},
			Status: batchv1.JobStatus{
				Succeeded: 1,
			},
		}

		kube := k8s.NewFakeKube(testJob)
		job := NewJob(appCtx, kube)

		state, err := job.GetState(context.Background())
		g.Expect(err).To(o.Succeed())
		g.Expect(state).To(o.Equal(Done))
	})

	t.Run("state_not_found_for_dry_run_job", func(t *testing.T) {
		testJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
								Args:  []string{"deploy", "--dry-run"},
							},
						},
					},
				},
			},
			Status: batchv1.JobStatus{
				Active: 1,
			},
		}

		kube := k8s.NewFakeKube(testJob)
		job := NewJob(appCtx, kube)

		state, err := job.GetState(context.Background())
		g.Expect(err).To(o.Succeed())
		// Dry-run jobs are considered NotFound
		g.Expect(state).To(o.Equal(NotFound))
	})

	t.Run("state_unknown", func(t *testing.T) {
		testJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"type": "installer-job." + annotations.RepoURI,
				},
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
								Args:  []string{"deploy"},
							},
						},
					},
				},
			},
			Status: batchv1.JobStatus{
				// No active, failed, or succeeded
			},
		}

		kube := k8s.NewFakeKube(testJob)
		job := NewJob(appCtx, kube)

		state, err := job.GetState(context.Background())
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("unknown job state"))
		g.Expect(state).To(o.Equal(JobState(-1)))
	})
}

// TestErrJobNotFound tests the error constant
func TestErrJobNotFound(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(ErrJobNotFound).ToNot(o.BeNil())
	g.Expect(ErrJobNotFound.Error()).To(o.Equal("job not found"))
	g.Expect(errors.Is(ErrJobNotFound, ErrJobNotFound)).To(o.BeTrue())
}
