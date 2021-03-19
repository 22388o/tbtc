package block

import (
	"context"
	"time"

	"github.com/ipfs/go-log"
	"github.com/keep-network/tbtc/relay/pkg/btc"
	"github.com/keep-network/tbtc/relay/pkg/chain"
)

const (
	// Size of the headers queue.
	headersQueueSize = 50

	// Maximum size of processed headers batch.
	headersBatchSize = 5

	// Maximum time for which the pulling process will wait for a single header
	// to be delivered by the headers queue.
	headerTimeout = 1 * time.Second

	// Block duration of a Bitcoin difficulty epoch.
	difficultyEpochDuration = 2016

	// Duration for which the forwarder should rest after performing
	// a push action.
	forwarderSleepTime = 45 * time.Second
)

var logger = log.Logger("relay-block-forwarder")

// Forwarder takes blocks from the Bitcoin chain and forwards them to the
// given host chain.
type Forwarder struct {
	btcChain  btc.Handle
	hostChain chain.Handle

	processedHeaders int

	headersQueue chan *btc.Header
	errChan      chan error
}

// RunForwarder creates an instance of the block forwarder and runs its
// processing loop. The lifecycle of the forwarder loop can be managed
// using the passed context.
func RunForwarder(
	ctx context.Context,
	btcChain btc.Handle,
	hostChain chain.Handle,
) *Forwarder {
	forwarder := &Forwarder{
		btcChain:     btcChain,
		hostChain:    hostChain,
		headersQueue: make(chan *btc.Header, headersQueueSize),
		errChan:      make(chan error, 1),
	}

	go forwarder.loop(ctx)

	return forwarder
}

func (f *Forwarder) loop(ctx context.Context) {
	logger.Infof("running forwarder")

	for {
		select {
		case <-ctx.Done():
			logger.Infof("forwarder context is done")
			return
		default:
			logger.Infof("pulling new headers from queue")

			headers := f.pullHeadersFromQueue(ctx)
			if len(headers) == 0 {
				continue
			}

			logger.Infof(
				"pushing [%v] header(s) to host chain",
				len(headers),
			)

			f.pushHeadersToChain(headers)

			logger.Infof(
				"suspending forwarder for [%v]",
				forwarderSleepTime,
			)

			// Sleep for a while to achieve a limited rate.
			select {
			case <-time.After(forwarderSleepTime):
			case <-ctx.Done():
			}
		}
	}
}

// pullHeadersFromQueue waits until we have `headersBatchSize` headers from
// the queue or until the queue fails to yield a header for
// `headerTimeout` duration.
func (f *Forwarder) pullHeadersFromQueue(ctx context.Context) []*btc.Header {
	headers := make([]*btc.Header, 0)

	headerTimer := time.NewTimer(headerTimeout)
	defer headerTimer.Stop()

	for len(headers) < headersBatchSize {
		logger.Debugf("waiting for new header appear on queue")

		select {
		case header := <-f.headersQueue:
			logger.Debugf("got header (%v) from queue", header.Height)

			headers = append(headers, header)

			// Stop the timer. In case it already expired, drain the channel
			// before performing reset.
			if !headerTimer.Stop() {
				<-headerTimer.C
			}
			headerTimer.Reset(headerTimeout)
		case <-headerTimer.C:
			if len(headers) > 0 {
				logger.Debugf(
					"new header did not appear in the given timeout; " +
						"returning headers pulled so far",
				)
				return headers
			}

			logger.Debugf(
				"new header did not appear in the given timeout; " +
					"resetting timer as no headers have been pulled so far",
			)

			// Timer expired and channel is drained so one can reset directly.
			headerTimer.Reset(headerTimeout)
		case <-ctx.Done():
			return headers
		}
	}

	return headers
}

func (f *Forwarder) pushHeadersToChain(headers []*btc.Header) {
	if len(headers) == 0 {
		return
	}

	startDifficulty := headers[0].Height % difficultyEpochDuration
	endDifficulty := headers[len(headers)-1].Height % difficultyEpochDuration

	if startDifficulty == 0 {
		// we have a difficulty change first
		// TODO: implementation
	} else if startDifficulty > endDifficulty {
		// we span a difficulty change
		// TODO: implementation
	} else {
		// no difficulty change
		// TODO: implementation
	}

	f.processedHeaders += len(headers)
	if f.processedHeaders >= headersBatchSize {
		newBestHeader := headers[len(headers)-1]
		f.updateBestHeader(newBestHeader)
		f.processedHeaders = 0
	}
}

func (f *Forwarder) updateBestHeader(header *btc.Header) {
	// TODO: implementation
}

// ErrChan returns the error channel of the forwarder. Once an error
// appears here, the forwarder loop is immediately terminated.
func (f *Forwarder) ErrChan() <-chan error {
	return f.errChan
}
