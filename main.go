package main

import (
	"fmt"
	"os"

	// "strings"

	// Temporary for testing output to file through string builder. Remove when done.
	"github.com/go-audio/wav"
	//"golang.org/x/exp/slices" // Import? or just manually do it?
	"gonum.org/v1/gonum/dsp/fourier" // "gonum"
	"gonum.org/v1/gonum/dsp/window"  // "gonum
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

	// fmt.Println(audiobuffer.Format.NumChannels)
	// fmt.Println(audiobuffer.Format.SampleRate) // Sample Rate will be found in aeneas? Or we find an "ideal" value independently? Or we need to match rates etc, to the provided audio?
	//fmt.Println(audiobuffer.Data)              // sin wave data
	audiofile.Close()

	// fmt.Println(audiobuffer.Data) // sin wave data

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

	// fmt.Println(len(signal))

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

	// fmt.Println(signal64)

	// End Normalization
	// Start Framing the Signal

	// How much overlap is desired
	// Slice of slices
	// Orginial signal slice
	// Each new slice to put in the slice of slices

	// 44100(samplingRate) * 0.03(milliseconds) = 1323 (frame size)

	// Sampling Rate = 44.1kHz
	// Frame Length = 20-30ms probably is a good choice

	var sampleRate = 44100.0
	var frameLength = 0.03
	var frameOverlap = 0.5
	var frameSize = int(sampleRate * frameLength)                                        // int: can't have a fraction of a sample
	var frameStep = int(sampleRate * frameLength * frameOverlap)                         // int: indexes are whole numbers
	var numFrames = (float64(len(signal64)) / (sampleRate * frameLength)) / frameOverlap // +1 if partial
	if numFrames != float64(int(numFrames)) {                                            // If there is a partial frame truncate and add 1, then handle the the partial/tail frame.
		numFrames = float64(int(numFrames))
	}
	fmt.Println("signal64: ", len(signal64))
	fmt.Println("frameSize: ", frameSize)
	fmt.Println("frameStep: ", frameStep)
	fmt.Println("numFrames: ", numFrames)

	// So an overlap variable might be to multiply by the frame length by desired overlap. So for a 50% overlap, multiply framelength by 0.5. So, for [i] would be for [i+0.5*frameLength]

	//for (i := 0; i < numFrames; i++) {

	// frames := slices.Slice(signal64, 1323, 1323) // "slices"

	frames := make([][]float64, int(numFrames))
	for i := 0; i < int(numFrames); i++ {
		start := i * frameStep
		end := start + frameSize
		if end > len(signal64) {
			end = len(signal64)
		}
		frames[i] = signal64[start:end]
	}
	///////////////////////////////////////////////////////////////////////
	// Testing Output to File to use in online array visualizer
	// Convert the slice of slices to a string representation
	// var sb strings.Builder
	// for _, row := range frames {
	// 	//fmt.Fprintf(&sb, "%v\n", row)
	// 	fmt.Fprintf(&sb, "%s\n", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(row)), ","), "[]"))
	// }
	// data := sb.String()

	// os.WriteFile("frames.txt", []byte(fmt.Sprint(data)), 0644)
	// fmt.Println("Finished Writing to File")
	///////////////////////////////////////////////////////////////////////
	// Count the samples per frame then add them all up and make sure they equal the starting length of the signal.
	// var count int
	// for i := 0; i < len(frames); i++ {
	// 	// fmt.Println(len(frames[i]))
	// 	count += len(frames[i])
	// 	if i == len(frames)-2 {
	// 		fmt.Println("Count: ", count)
	// 		fmt.Println("2ndtolast: ", len(frames[i]))
	// 	}
	// 	if i == len(frames)-1 {
	// 		fmt.Println("Count: ", count)
	// 		fmt.Println("lastframe: ", len(frames[i]))
	// 	}
	// 	// Print the first overlap index of each frame
	// 	if i != 0 {
	// 		fmt.Println("Previous Frame: ", frames[i-1][660:663])
	// 		fmt.Println("Current Frame : ", frames[i][0:3])
	// 	}
	///////////////////////////////////////////////////////////////////////
	// Check to see if the last frames align. Hopefully this means that I've gathered all the data correctly.
	// fmt.Println(len(frames))
	// fmt.Println(frames[796][1257])
	// fmt.Println(signal64[len(signal64)-1])
	///////////////////////////////////////////////////////////////////////

	// End Framing the Signal
	// Start Windowing the Signal (Hamming Window)
	// https://pkg.go.dev/gonum.org/v1/gonum/dsp/window#example-Hamming
	// I believe the windowing function is applied inplace, so I need to create the slice of slices beforehand otherwise the windowing would change the values, which is a problem since there is going to be overlap and the values will need to be resued for multiple frames and we don't want it to get "windowed" multiple times?

	// Have to copy the frames because I need both the original and windowed frames for the FFT.

	goingHam := make([][]float64, len(frames))
	for i := 0; i < len(frames); i++ {
		goingHam[i] = append(goingHam[i], frames[i]...)
	}

	// Maybe run a quick test to see if they are equivalent, then remove it.
	// eg, for i in range, if frames[i] == goingHam[i] then print "same" else print "different"
	// HAVE NOT YET RUN THIS TEST

	for i := 0; i < len(goingHam); i++ {
		window.Hamming(goingHam[i]) // Changes data in place according to documentation
	}
	// Now the overlap-add step. https://en.wikipedia.org/wiki/Overlap%E2%80%93add_method
	// recombining the frames accounting for the overlap.

	// - FFT then overlap-add
	// - Overlap-add method
	// There doesn't seem to be a library for this, so, like framing, do it manually.
	// -

	// So even though the slice of slices are independent from each other and so there isn't an issue with the windowing function changing the values of the other frames, I still need to account for the overlap?

	// End Windowing the Signal

	// The DCT (or mdct) happens last

	// The FFT returns complex numbers, but Do I need a complex input?

	complexFrames := make([][]complex128, len(frames))
	fft := fourier.NewFFT(len(frames))
	for i := 0; i < len(frames); i++ {
		complexFrames[i] = make([]complex128, len(frames[i]))
		for j := 0; j < len(frames[i]); j++ {
			complexFrames[i][j] = complex(frames[i][j], 0)
		}
		fft.Coefficients(complexFrames[i], goingHam[i]) // The first argument should be the signal before windowing. The second argument is the windowed bit.
		// This means I have to go back to the windowing and do it without changing the original signal.
	}

}
