package composeFile

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Ports []interface{}

type ServiceSettings map[interface{}]map[string]interface{}
type Compose struct {
	Services ServiceSettings `yaml:"services"`
}
type Handler struct {
	port string
}
type IHandler interface {
	Ports(composeFilePaths []string) (prettyPrintPorts string, err error)
}

func NewPortHandler(port string) IHandler {
	return &Handler{
		port: port,
	}
}

func (h Handler) Ports(composeFilePaths []string) (prettyPrintPorts string, err error) {
	composeFilesContent, err := h.composeContent(composeFilePaths)
	if err != nil {
		return "", err
	}
	// get list of all current port ranges
	// TODO: add support for log syntax:
	// https://docs.docker.com/compose/compose-file/compose-file-v3/
	occupiedPortRanges, err := h.occupiedPortRanges(composeFilesContent)
	if err != nil {
		return "", err
	}

	// create a set of existing ports
	hostPorts, err := h.hostPorts(occupiedPortRanges, err)
	// get next available ports
	nextPorts := h.nextAvailablePorts(hostPorts)
	if h.port == "" {
		return fmt.Sprintf("next available ports %v", nextPorts), nil
	}
	// if --port is provided filter based on it
	filteredPorts := h.filterPorts(nextPorts)
	return fmt.Sprintf("requested ports: %v", filteredPorts), nil
}

func (h Handler) filterPorts(nextPorts []int64) []int64 {
	portRegex := ""
	for _, char := range h.port {
		if unicode.IsDigit(char) {
			portRegex = fmt.Sprintf("%s%s", portRegex, string(char))
		} else {
			portRegex = fmt.Sprintf("%s[0-9]+", portRegex)
		}
	}
	r, _ := regexp.Compile(portRegex)
	requestedPorts := []int64{}
	for _, p := range nextPorts {
		if r.MatchString(strconv.FormatInt(p, 10)) {
			requestedPorts = append(requestedPorts, p)
		}
	}
	return requestedPorts
}

func (h Handler) nextAvailablePorts(hostPorts map[int64]struct{}) []int64 {
	hostPortsSlice := []int64{}
	for key, _ := range hostPorts {
		hostPortsSlice = append(hostPortsSlice, key)
	}
	sort.Slice(hostPortsSlice, func(i, j int) bool {
		return hostPortsSlice[i] < hostPortsSlice[j]
	})
	// get the next available ports
	nextPorts := []int64{}
	for i := 0; i < len(hostPortsSlice); i++ {
		if i == len(hostPortsSlice)-1 {
			nextPorts = append(nextPorts, hostPortsSlice[i]+1)
			break
		}
		if hostPortsSlice[i]+1 != hostPortsSlice[i+1] {
			nextPorts = append(nextPorts, hostPortsSlice[i]+1)
		}
	}
	return nextPorts
}

func (h Handler) hostPorts(occupiedPortRanges []string, err error) (map[int64]struct{}, error) {
	hostPorts := map[int64]struct{}{}
	for _, p := range occupiedPortRanges {
		if strings.Contains(p, ":") {
			dPorts := strings.Split(p, ":")
			if len(dPorts) == 2 {
				hostPort := dPorts[0]
				hostPorts, err = h.currentHostPorts(hostPort, hostPorts)
			}
			if len(dPorts) == 3 {
				hostPort := dPorts[1]
				hostPorts, err = h.currentHostPorts(hostPort, hostPorts)
			}
		}
	}
	return hostPorts, err
}

func (h Handler) occupiedPortRanges(composeFilesContent []Compose) ([]string, error) {
	occupied := []string{}
	for _, c := range composeFilesContent {
		for _, settings := range c.Services {
			if p, ok := settings["ports"]; ok {
				var ports, ok = p.([]interface{})
				if !ok {
					return nil, fmt.Errorf("no ports")
				}
				for _, port := range ports {
					occupied = append(occupied, fmt.Sprintf("%v", port))
				}
			}
		}
	}
	return occupied, nil
}

func (h Handler) composeContent(composeFilePaths []string) ([]Compose, error) {
	composeFilesContent := []Compose{}
	for _, f := range composeFilePaths {
		compose := Compose{}
		bytes, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(bytes, &compose)
		if err != nil {
			return nil, err
		}
		composeFilesContent = append(composeFilesContent, compose)
	}
	return composeFilesContent, nil
}

func (h Handler) currentHostPorts(hostPort string, hostPorts map[int64]struct{}) (updatedHostPorts map[int64]struct{}, err error) {
	updatedHostPorts = hostPorts
	if strings.Contains(hostPort, "-") {
		hostPortRange := strings.Split(hostPort, "-")
		startRange, err := strconv.ParseInt(hostPortRange[0], 10, 64)
		if err != nil {
			return nil, err
		}
		endRange, err := strconv.ParseInt(hostPortRange[1], 10, 64)
		if err != nil {
			return nil, err
		}
		for i := startRange; i <= endRange; i++ {
			hostPorts[i] = struct{}{}
		}
	} else if hostPort != "" {
		hp, err := strconv.ParseInt(hostPort, 10, 64)
		if err != nil {
			return nil, err
		}
		hostPorts[hp] = struct{}{}
	}
	return hostPorts, nil
}
