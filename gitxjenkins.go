package main

import (
    "github.com/google/go-github/github"
    "fmt"
    "github.com/bndr/gojenkins"
    "github.com/anthonydahanne/stash"
    "net/url"
    "strings"
    "bytes"
    "gopkg.in/xmlpath.v1"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "sort"
    "time"
    "log"
    "html/template"
    "os"
    "bufio"
)

type Jenkins struct {
    Name     string
    Url      string
    Username string
    Password string
}

type Github struct {
    Organization string
    Username     string
    Password     string
}

type Stash struct {
    Prefix   string
    Url      string
    Username string
    Password string
}

type Configuration struct {
    Jenkins []Jenkins
    Github  []Github
    Stash   []Stash
}

type Job struct {
    JenkinsName string
    Name        string
    Url         string
    ScmUrl      string
}

type GitRepository struct {
    Project string
    ProjectUrl string
    Name    string
    Url     string
    ScmUrls []string
    Jobs    []Job
}

type ByProjectAndName []GitRepository

func (a ByProjectAndName) Len() int { return len(a) }
func (a ByProjectAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByProjectAndName) Less(i, j int) bool { return a[i].Project + a[i].Name < a[j].Project + a[j].Name}

func main() {

    fmt.Printf("Loading configuration...\n")
    configuration, err := readConfiguration();
    if err != nil {
        fmt.Printf("error: %v", err)
        return
    }
    fmt.Printf("Finished loading configuration.\n")


    var allJobs []Job
    var allRepos[]GitRepository

    for _, jenkins := range configuration.Jenkins {
        fmt.Printf("Loading jobs from %v...\n", jenkins.Name)
        resultJobs, err := jenkinsParser(jenkins.Name, jenkins.Url, jenkins.Username, jenkins.Password)
        if err != nil {
            fmt.Printf("error: %v", err)
            return
        }
        fmt.Printf("Finished loading jobs from %v\n", jenkins.Name)
        allJobs = append(allJobs, resultJobs...)
    }

    for _, github := range configuration.Github {
        fmt.Printf("Loading repos from %v...\n", github.Organization)
        resultRepos, err := githubParser(github.Organization, github.Username, github.Password)
        if err != nil {
            fmt.Printf("error: %v", err)
            return
        }
        fmt.Printf("Finished loading repos from %v\n", github.Organization)
        allRepos = append(allRepos, resultRepos...)
    }


    for _, stash := range configuration.Stash {
        fmt.Printf("Loading repos from %v...\n", stash.Url)
        resultRepos, err := stashParser(stash.Url, stash.Prefix, stash.Username, stash.Password)
        if err != nil {
            fmt.Printf("error: %v", err)
            return
        }
        fmt.Printf("Finished loading repos with project prefixed with %v,  from %v", stash.Prefix, stash.Url)
        allRepos = append(allRepos, resultRepos...)
    }

    sort.Sort(ByProjectAndName((allRepos)))

    attachJobsToRepos(allRepos, allJobs)

    printResultsAsText(allRepos)

    printResultsAsHtml(allRepos, configuration)


}

func printResultsAsText(allRepos []GitRepository) {
    for _, repo := range allRepos {
        if (len(repo.Jobs) > 0) {
            jobsAsStrings := ""
            for _, job := range repo.Jobs {
                if (len(jobsAsStrings) > 0) {
                    jobsAsStrings += " and "
                }
                jobsAsStrings += job.Name + " at " + job.Url
            }
            fmt.Printf("%v/%v is built by %v \n", repo.Project, repo.Name, jobsAsStrings)
        } else {
            fmt.Printf("%v/%v is an orphan repo\n", repo.Project, repo.Name)
        }
    }
}


func printResultsAsHtml(allRepos []GitRepository, configuration Configuration) {
    currentTime := time.Now().Format(time.RFC3339)


    check := func(err error) {
        if err != nil {
            log.Fatal(err)
        }
    }
    t, err := template.ParseFiles("./template.html")

    type ConfigElement struct {
        Name    string
        Url string
    }
    configElements := []ConfigElement{}

    for _, jenkins := range configuration.Jenkins {
        configElements = append(configElements, ConfigElement{jenkins.Username + " @ " + jenkins.Name, jenkins.Url})
    }
    for _, github := range configuration.Github {
        configElements = append(configElements, ConfigElement{github.Organization, "https://github.com/" + github.Organization })
    }
    for _, stash := range configuration.Stash {
        configElements = append(configElements, ConfigElement{stash.Prefix, stash.Url})
    }

    for _, repo := range allRepos {
        if (len(repo.Jobs) > 0) {
            jobsAsStrings := ""
            for _, job := range repo.Jobs {
                if (len(jobsAsStrings) > 0) {
                    jobsAsStrings += " and "
                }
                jobsAsStrings += job.Name + " at " + job.Url
            }
        }
    }

    data := struct {
        Title          string
        Date           string
        ConfigElements []ConfigElement
        ReposAndJobs   []GitRepository

    }{
        Title: "GitXJenkins report",
        Date: currentTime,
        ConfigElements: configElements,
        ReposAndJobs: allRepos,
    }


    f, err := os.Create("output.html")
    check(err)
    defer f.Close()

    w := bufio.NewWriter(f)
    err = t.Execute(w, data)
    check(err)

    w.Flush()

}


func attachJobsToRepos(allRepos []GitRepository, allJobs []Job) {
    for i, repo := range allRepos {
        for _, job := range allJobs {
            for _, repoScmUrl := range repo.ScmUrls {
                if (job.ScmUrl == repoScmUrl) {
                    allRepos[i].Jobs = append(allRepos[i].Jobs, job)
                }
            }
        }
    }
}

func jenkinsParser(jenkinsName string, jenkinsURL string, username string, password string) ([]Job, error) {

    var resultJobs []Job

    var jenkins *gojenkins.Jenkins
    var err error

    if (username != "") {
        jenkins, err = gojenkins.CreateJenkins(jenkinsURL, username, password).Init()
    } else {
        jenkins, err = gojenkins.CreateJenkins(jenkinsURL).Init()
    }
    if err != nil {
        return nil, err
    }

    fmt.Printf("%s: Connected\n", jenkinsName)

    status, err := jenkins.Poll();
    if(status != 200) {
        log.Fatalf("Jenkins replied with status %v ; check your credentials for %v", status, jenkinsURL)
        return nil, err
    }

    fmt.Printf("%s: Polled\n", jenkinsName)

    jobs, err := jenkins.GetAllJobs();
    if err != nil {
        return nil, err
    }

    fmt.Printf("%s: GetAllJobs done (%d jobs)\n", jenkinsName, len(jobs))

    path := xmlpath.MustCompile("//scm/userRemoteConfigs/hudson.plugins.git.UserRemoteConfig/url")

    for _, job := range jobs {
        jobConfig, err := job.GetConfig();
        if err != nil {
            fmt.Printf("error: %v", err)
            continue
        }

        b := bytes.NewBufferString(jobConfig)

        root, err := xmlpath.Parse(b)
        if err != nil {
            fmt.Printf("error: %v", err)
            continue
        }
        if scmUrl, ok := path.String(root); ok {
            resultJobs = append(resultJobs, Job{JenkinsName: jenkinsName, Name:job.GetName(), Url:job.GetDetails().URL, ScmUrl:scmUrl})
        }
    }
    return resultJobs, nil
}

func stashParser(stashBaseUrl string, projectPrefix string, username string, password string) ([]GitRepository, error) {

    var resultGitRepositories []GitRepository

    stashUrl, err := url.Parse(stashBaseUrl)
    if err != nil {
        return resultGitRepositories, err
    }
    stashClient := stash.NewClient(username, password, stashUrl)
    repository, err := stashClient.GetRepositories()
    if err != nil {
        return resultGitRepositories, err
    }
    for _, element := range repository {
        if ( strings.HasPrefix(element.Project.Key, projectPrefix)) {
            var urls []string
            urls = append(urls, element.SshUrl())
            repoUrl := stashBaseUrl + "/projects/" + element.Project.Key + "/repos/" + element.Name + "/browse"
            resultGitRepositories = append(resultGitRepositories, GitRepository{Project: element.Project.Key, Name: element.Name, Url: repoUrl, ScmUrls:urls})
        }
    }
    return resultGitRepositories, nil
}


func githubParser(githubOrganization string, username string, password string) ([]GitRepository, error) {
    var resultGitRepositories []GitRepository

    client := github.NewClient(nil)
    orgs, _, _ := client.Repositories.List(githubOrganization, nil)

    for _, element := range orgs {
        var urls []string
        urls = append(urls, *element.GitURL)
        urls = append(urls, *element.CloneURL)
        url := strings.Replace(*element.URL, "api.", "", -1)
        url = strings.Replace(url, "repos/", "", -1)
        resultGitRepositories = append(resultGitRepositories, GitRepository{Project:githubOrganization, ProjectUrl:"https://github.com/" + githubOrganization, Name:*element.Name, Url:url, ScmUrls:urls})
    }
    return resultGitRepositories, nil
}

func readConfiguration() (Configuration, error) {
    t := Configuration{}
    dat, err := ioutil.ReadFile("config.yml")
    if err != nil {
        return t, err
    }
    err = yaml.Unmarshal(dat, &t)
    if err != nil {
        return t, err
    }
    return t, nil
}
