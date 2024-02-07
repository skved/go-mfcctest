package main

import (
	"fmt"
	"os"

	"github.com/go-audio/wav"
	//"golang.org/x/exp/slices" // Import? or just manually do it?
)

// Load audio file
// Apply Pre-emphasis
// Amplify the high frequencies
// Frame the Signal
// Create overlapping segments (overlapping reduces artifacting?)
// Apply a Window Function - Like hamming
// Each frame get "windowed" by multiplying it with a window function
// Apply FFT to each windowed frame
// Converts from time domain to frequency domain
// Apply Triangular Filtering - A type of mel filter, so still spaced on the mel scale.
// The filter banks (triangular filtering) divide the signal's frequency spectrum into multiple frequency bands.
// So that each band can be analyzed seperately.
// Apply a logarithm to the filter bank energies
//Filter bank energies - computed by summing the magnitudes of the FFt output within each of the triangular filters
// Take the DCT - Discrete cosine transform
// Take the DCT of the list of mel log powers
// The MFCCs are the amplitudes of the resulting spectrum

const preEmphasis = 0.95 // Modify coefficient as needed

func main() {
	// Variables needed?
	audiofile, err := os.Open("test.wav") // "os"
	if err != nil {
		// need to put this in a function so that I can have a proper return type
		// os.open is forcing an error return type
		print("Error opening file")
	}

	decoder := wav.NewDecoder(audiofile) // "wav" // decoder // This should automatically
	if decoder == nil {
		print("Error decoding file")
	}
	audiobuffer, err := decoder.FullPCMBuffer() // "wav"
	if err != nil {
		print("why is everything broken")
	}
	// FullPCMBuffer takes a pointer to a decoder
	// returns a buffer
	// PCMBuffer is more efficient

	fmt.Println(audiobuffer.Format.NumChannels)
	fmt.Println(audiobuffer.Format.SampleRate) // Sample Rate will be found in aeneas? Or we find an "ideal" value independently? Or we need to match rates etc, to the provided audio?
	//fmt.Println(audiobuffer.Data)              // sin wave data
	audiofile.Close()

	fmt.Println(audiobuffer.Data) // sin wave data

	// End Signal Loading
	// Start Pre Emphasis

	signal := audiobuffer.Data

	// Convert to float
	// SEE IF THERE IS A BETTER WAY
	// THERE ARE 500000 values!!
	signal64 := make([]float64, len(signal))
	for i := 0; i < len(signal); i++ {
		signal64[i] = float64(signal[i])
	}
	// End
	for i := 1; i < len(signal); i++ {
		signal64[i] = signal64[i] - preEmphasis*signal64[i-1]
	}

	fmt.Println(len(signal))

	// End Pre-Emphasis
	// Do we need to normalize the signal? Wikipedia seems to indicate that it is necessary for windowing.
	// Start Normalization
	// Find the biggest value in the signal and divide all other values by it.

	currentMax := signal64[0]
	if currentMax < 0 {
		currentMax = currentMax * -1
	}

	// I think this will accurately find the largest value but have someone double check for errors.
	for i := 1; i < len(signal64); i++ {
		next := signal64[i]
		if next < 0 {
			next = next * -1
		}
		if next > currentMax {
			currentMax = next
		}
	}
	absMax := currentMax

	for i := 0; i < len(signal); i++ {
		signal64[i] = signal64[i] / absMax
	}

	fmt.Println(signal64)

	// End Normalization
	// Start Framing the Signal

	// How much overlap is desired
	// Slice of slices
	// Orginial signal slice
	// Each new slice to put in the slice of slices

	// 44100(samplingRate) * 0.03(milliseconds) = 1323 (frame size)

	var numFrames = len(signal64) / 1323 // +1 partial
	fmt.Println(numFrames)

	// frames := slices.Slice(signal64, 1323, 1323) // "slices"

	// End Framing the Signal

}