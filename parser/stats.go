package parser

import (
	"context"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// OutputType is an enum of Docker Output Processing Types.
type OutputType int

const (
	// DockerPull is the "docker pull" output processing type.
	DockerPull OutputType = iota
	// DockerPush is the "docker push" output processing type.
	DockerPush
)

// ProcessingStats contains all the statistics of the running layers.
type ProcessingStats struct {
	Last   StatusChange
	Format OutputType

	// Common
	AlreadyExists []string

	// Pull
	PullingFSLayer    []string
	VerifyingChecksum []string
	DownloadComplete  []string
	PullComplete      []string
	LayerInprogress   []string
	LayerList         []string

	// Push
	Preparing []string
	Waiting   []string
	Pushed    []string

	// Internal
	m sync.RWMutex
}

// NewProcessingStats returns a new ProcessingStats object.
func NewProcessingStats() *ProcessingStats {
	return &ProcessingStats{
		m:      sync.RWMutex{},
		Format: DockerPull,
	}
}

// Pulling fs layer
// Verifying Checksum
// Download complete
// Pull complete
// test-container: Pulling from koshatul/ci

func (p *ProcessingStats) addLayerInprogress(layer string) {
	p.addLayer(layer)

	p.m.Lock()
	defer p.m.Unlock()

	for _, n := range p.LayerInprogress {
		if strings.Compare(layer, n) == 0 {
			return
		}
	}

	p.LayerInprogress = append(p.LayerInprogress, layer)
}

func (p *ProcessingStats) removeLayerInprogress(layer string) {
	p.m.Lock()
	defer p.m.Unlock()

	layers := []string{}

	for _, l := range p.LayerInprogress {
		if strings.Compare(layer, l) == 0 {
			continue
		}

		layers = append(layers, l)
	}

	p.LayerInprogress = layers
}

func (p *ProcessingStats) addLayer(layer string) {
	p.m.Lock()
	defer p.m.Unlock()

	for _, n := range p.LayerList {
		if strings.Compare(layer, n) == 0 {
			return
		}
	}

	p.LayerList = append(p.LayerList, layer)
}

// SetFormat set the current output format for the docker push/pull.
func (p *ProcessingStats) SetFormat(f OutputType) {
	p.m.Lock()
	defer p.m.Unlock()

	p.Format = f
}

func (p *ProcessingStats) appendWithLock(field *[]string, chg StatusChange) {
	p.m.Lock()
	*field = append(*field, chg.LayerName)
	p.m.Unlock()
}

// Run processes the queue until there are no messages left.
//nolint: funlen // not worth splitting up switch, it's more readable here.
func (p *ProcessingStats) Run(ctx context.Context, processingQueue chan StatusChange) {
	for chg := range processingQueue {
		if chg.LayerName != "" {
			p.Last = chg
		}

		switch chg.Status {
		// Internal
		case "PRINT":
			p.Print()
		// Push
		case "Preparing":
			p.appendWithLock(&p.Preparing, chg)
			p.addLayerInprogress(chg.LayerName)
			p.Print()
		case "Waiting":
			p.appendWithLock(&p.Waiting, chg)
			p.addLayer(chg.LayerName)
			p.Print()
		case "Pushed":
			p.appendWithLock(&p.Pushed, chg)
			p.addLayer(chg.LayerName)
			p.removeLayerInprogress(chg.LayerName)
			p.Print()
		case "Layer already exists":
			p.appendWithLock(&p.AlreadyExists, chg)
			p.addLayer(chg.LayerName)
			p.removeLayerInprogress(chg.LayerName)
			p.Print()
			// Pull
		case "Pulling fs layer":
			p.appendWithLock(&p.PullingFSLayer, chg)
			p.addLayerInprogress(chg.LayerName)
			p.Print()
		case "Verifying Checksum":
			p.appendWithLock(&p.VerifyingChecksum, chg)
			p.addLayer(chg.LayerName)
			p.Print()
		case "Download complete":
			p.appendWithLock(&p.DownloadComplete, chg)
			p.addLayer(chg.LayerName)
			p.Print()
		case "Pull complete":
			p.appendWithLock(&p.PullComplete, chg)
			p.addLayer(chg.LayerName)
			p.removeLayerInprogress(chg.LayerName)
			p.Print()
			// Common
		case "Already Exists":
			p.appendWithLock(&p.AlreadyExists, chg)
			p.addLayer(chg.LayerName)
			p.removeLayerInprogress(chg.LayerName)
			p.Print()
		}
	}
}

// Print displays the current state of the processing.
//nolint: lll // long log output, not worth splitting over lines.
func (p *ProcessingStats) Print() {
	p.m.RLock()
	defer p.m.RUnlock()

	switch p.Format {
	case DockerPull:
		logrus.Infof(
			"Last:[%s: %s]; Pulling FS Layer:%d; Verifying Complete:%d; Download Complete:%d; Pull Complete:%d; InProgress:%d; Total:%d",
			p.Last.LayerName, p.Last.Status,
			len(p.PullingFSLayer),
			len(p.VerifyingChecksum),
			len(p.DownloadComplete),
			len(p.PullComplete),
			len(p.LayerInprogress),
			len(p.LayerList),
		)
	case DockerPush:
		logrus.Infof(
			"Last:[%s: %s]; Preparing:%d; Waiting:%d; Already Exists:%d; Pushed:%d; InProgress:%d; Total:%d",
			p.Last.LayerName, p.Last.Status,
			len(p.Preparing),
			len(p.Waiting),
			len(p.AlreadyExists),
			len(p.Pushed),
			len(p.LayerInprogress),
			len(p.LayerList),
		)
	}
}
