# Working with Resources

*This document is aimed at users who have worked through some of the examples,
and who want to learn more about using kubectl to manage resources such
as pods and services.  Users who want to access the REST API directly,
and developers who want to extend the kubernetes API should 
refer to the [api conventions](api-conventions.md) and
the [api document](api.md).*

## Resources are Automatically Modified
When you create a resource such as pod, and then retrieve the created
resource, a number of the fields of the resource are added.
You can see this at work in the following example:
```
$ cat > original.yaml <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: foo
spec:
  containers:
    - name: foo
      image: busybox
  restartPolicy: Never
EOF
$ kubectl create -f original.yaml
pods/original
$ kubectl get pods/original -o yaml > current.yaml
pods/original
$ wc -l original.yaml current.yaml
      51 current.yaml
       9 original.yaml
      60 total
```
The resource we posted had only 9 lines, but the one we got back had 51 lines. 
If you `diff original.yaml current.yaml`, you can see the fields added to the pod.
The system adds fields in several ways:
  - Some fields are added synchronously with creation of the resource and some are set asynchronously.
    - For example: `metadata.uid` is set synchronously.  (Read more about [metadata](api-conventions.md#metadata)).
    - For example, `status.hostIP` is set only after the pod has been scheduled.  This often happens fast, but you may notice pods which do not have this set yet.  This is called Late Initialization.  (Read mode about [status](api-conventions.md#spec-and-status) and [late initialization](api-conventions.md#late-initialization) ).
  - Some fields are set to default values.  Some defaults vary by cluster and some are fixed for the API at a certain version.  (Read more about [defaulting](api-conventions.md#defaulting)).
    - For example, `spec.containers.imagePullPolicy` always defaults to `IfNotPresent` in api v1.
    - For example, `spec.containers.resources.limits.cpu` may be defaulted to  `100m` on some clusters, to some other value on others, and not defaulted at all on others.
The API will generally not modify fields that you have set; it just sets ones which were unspecified.

## <a name="finding_schema_docs"></a>Finding Documentation on Resource Fields
You can browse auto-generated API documentation at the [project website](http://kubernetes.io/third_party/swagger-ui/) or directly from your cluster, like this:
  - Run `kubectl proxy --api-prefix=/`
  - Go to `http://localhost:8001/swagger-ui` in your browser.
  - It should say "swagger" at the top-left.

Once there:
  - Click on "v1" and wait for it to expand.
  - Search for "pods", "services", "replicationcontrollers" or some other resource.
  - Click on that POST row for the matching resource.
  - Click on the words "Model".
  - You should see a list of all possible resource fields, starting with `v1.pods {...}`



[![Analytics](https://kubernetes-site.appspot.com/UA-36037335-10/GitHub/docs/working_with_resources.md?pixel)]()


[![Analytics](https://kubernetes-site.appspot.com/UA-36037335-10/GitHub/release-0.20.0/docs/working_with_resources.md?pixel)]()
