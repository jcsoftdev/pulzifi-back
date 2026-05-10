import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import {
  ArrowRight,
  Brain,
  CheckCircle2,
  FileText,
  Library,
  Mic,
  Play,
  Sparkles,
  Users,
  VolumeX,
  Zap,
} from 'lucide-react'
import type { ReactNode } from 'react'
import { CompareBars } from './compare-bars'
import { WaveBars } from './wave-bars'

export default function LectureAiLandingPage() {
  const v = '2'

  const isV2 = v === '2'

  return (
    <div className="min-h-screen bg-background text-foreground selection:bg-primary/20 font-sans">
      {/* Navigation */}
      <nav className="border-b border-border/40 backdrop-blur-md sticky top-0 z-50 bg-background/80 supports-backdrop-filter:bg-background/60">
        <div className="container mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-primary/10 p-2 rounded-lg">
              <Brain className="w-5 h-5 text-primary" />
            </div>
            <span className="font-bold text-xl tracking-tight">LectureAI</span>
          </div>
          <div className="hidden md:flex items-center gap-8 text-sm font-medium text-muted-foreground">
            <a href="#features" className="hover:text-primary transition-colors">
              Features
            </a>
            <a href="#how-it-works" className="hover:text-primary transition-colors">
              How it Works
            </a>
            <a href="#testimonials" className="hover:text-primary transition-colors">
              Testimonials
            </a>
            <a href="#pricing" className="hover:text-primary transition-colors">
              Pricing
            </a>
          </div>
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              className="hidden sm:inline-flex text-muted-foreground hover:text-foreground"
            >
              Log in
            </Button>
            <Button className="font-semibold shadow-md shadow-primary/20">Get Started</Button>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative pt-24 pb-32 overflow-hidden">
        <div className="absolute inset-0 bg-[radial-linear(ellipse_at_top,var(--tw-linear-stops))] from-primary/10 via-background to-background pointer-events-none" />
        <div className="container mx-auto px-6 relative z-10">
          <div className="max-w-5xl mx-auto text-center flex flex-col items-center gap-8">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-secondary/50 border border-secondary text-secondary-foreground text-sm font-medium animate-in fade-in slide-in-from-bottom-4 duration-700">
              <Sparkles className="w-3.5 h-3.5" />
              <span>
                {isV2
                  ? 'Now with Real-time Citation Analysis'
                  : 'Introducing Smart Audio Isolation 2.0'}
              </span>
            </div>

            <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight text-foreground leading-[1.1] max-w-4xl">
              {isV2 ? (
                <>
                  Your personal AI tutor. <br />
                  <span className="text-transparent bg-clip-text bg-linear-to-r from-primary via-purple-500 to-blue-600">
                    From lecture to mastery.
                  </span>
                </>
              ) : (
                <>
                  Capture the knowledge. <br />
                  <span className="text-transparent bg-clip-text bg-linear-to-r from-primary via-blue-500 to-cyan-500">
                    Filter out the noise.
                  </span>
                </>
              )}
            </h1>

            <p className="text-xl text-muted-foreground max-w-2xl leading-relaxed">
              {isV2
                ? 'Turn messy lecture recordings into structured knowledge. Our AI filters irrelevant student chatter, enhances professor audio, and enriches transcripts with verified academic sources.'
                : 'The only recording app designed for the lecture hall. Automatically remove student interruptions, coughs, and background noise so you get a pristine recording of exactly what the professor taught.'}
            </p>

            <div className="flex flex-col sm:flex-row gap-4 w-full justify-center pt-6">
              <Button
                size="lg"
                className="h-14 px-8 text-lg rounded-full shadow-xl shadow-primary/20 hover:shadow-primary/30 transition-all hover:scale-105"
              >
                Start Recording Free
                <ArrowRight className="w-5 h-5 ml-2" />
              </Button>
              <Button
                size="lg"
                variant="outline"
                className="h-14 px-8 text-lg rounded-full bg-background/50 backdrop-blur-sm border-2 hover:bg-secondary/50"
              >
                <Play className="w-5 h-5 mr-2 fill-current" />
                Listen to Demo
              </Button>
            </div>

            {/* Social Proof */}
            <div className="mt-12 pt-8 border-t border-border/50 w-full max-w-3xl">
              <p className="text-sm text-muted-foreground mb-6 font-medium">
                TRUSTED BY STUDENTS AT TOP UNIVERSITIES
              </p>
              <div className="flex flex-wrap justify-center gap-x-12 gap-y-8 opacity-60 grayscale hover:grayscale-0 transition-all duration-500">
                {[
                  'MIT',
                  'Stanford',
                  'Oxford',
                  'Cambridge',
                  'ETH Zürich',
                ].map((uni) => (
                  <span key={uni} className="text-lg font-bold font-serif">
                    {uni}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* How it Works Steps */}
      <section id="how-it-works" className="py-24 bg-background">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              From chaos to clarity in three steps
            </h2>
            <p className="text-muted-foreground text-lg">
              LectureAI handles the heavy lifting so you can focus on understanding.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-12 relative">
            {/* Connector Line (Desktop) */}
            <div className="hidden md:block absolute top-12 left-0 w-full h-0.5 bg-linear-to-r from-transparent via-primary/20 to-transparent" />

            {/* Step 1 */}
            <div className="relative pt-8 text-center group">
              <div className="w-16 h-16 mx-auto bg-background border-2 border-primary/20 rounded-2xl flex items-center justify-center text-primary text-xl font-bold mb-6 relative z-10 group-hover:border-primary group-hover:scale-110 transition-all duration-300 shadow-lg shadow-primary/5">
                1
              </div>
              <h3 className="text-xl font-bold mb-3">Record</h3>
              <p className="text-muted-foreground">
                Hit the button at the start of class. Our app listens intelligently, separating the
                professor's voice from the room noise.
              </p>
            </div>

            {/* Step 2 */}
            <div className="relative pt-8 text-center group">
              <div className="w-16 h-16 mx-auto bg-background border-2 border-primary/20 rounded-2xl flex items-center justify-center text-primary text-xl font-bold mb-6 relative z-10 group-hover:border-primary group-hover:scale-110 transition-all duration-300 shadow-lg shadow-primary/5">
                2
              </div>
              <h3 className="text-xl font-bold mb-3">Process</h3>
              <p className="text-muted-foreground">
                Our AI engine transcribes the audio, filters out interruptions, and identifies key
                concepts and definitions in real-time.
              </p>
            </div>

            {/* Step 3 */}
            <div className="relative pt-8 text-center group">
              <div className="w-16 h-16 mx-auto bg-background border-2 border-primary/20 rounded-2xl flex items-center justify-center text-primary text-xl font-bold mb-6 relative z-10 group-hover:border-primary group-hover:scale-110 transition-all duration-300 shadow-lg shadow-primary/5">
                3
              </div>
              <h3 className="text-xl font-bold mb-3">Learn</h3>
              <p className="text-muted-foreground">
                Get a structured summary, search the transcript, and explore related academic papers
                to deepen your understanding.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Interactive Demo / UI Mockup */}
      <section className="py-12 bg-muted/30 border-y border-border/50">
        <div className="container mx-auto px-6">
          <div className="max-w-6xl mx-auto rounded-2xl border border-border shadow-2xl shadow-primary/20 bg-background overflow-hidden relative">
            <div className="absolute top-0 left-0 w-full h-1 bg-linear-to-r from-primary via-purple-500 to-blue-500" />
            <div className="grid lg:grid-cols-12 gap-0">
              {/* Sidebar */}
              <div className="lg:col-span-3 border-r border-border bg-muted/10 p-4 hidden lg:block">
                <div className="space-y-6">
                  <div>
                    <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
                      Recent Lectures
                    </h4>
                    <div className="space-y-1">
                      <div className="p-2 rounded-md bg-primary/10 text-primary font-medium text-sm flex items-center gap-2">
                        <Mic className="w-3.5 h-3.5" /> Adv. Calculus II
                      </div>
                      <div className="p-2 rounded-md hover:bg-muted text-muted-foreground text-sm flex items-center gap-2">
                        <CheckCircle2 className="w-3.5 h-3.5" /> Macroeconomics
                      </div>
                      <div className="p-2 rounded-md hover:bg-muted text-muted-foreground text-sm flex items-center gap-2">
                        <CheckCircle2 className="w-3.5 h-3.5" /> Org. Chemistry
                      </div>
                    </div>
                  </div>
                  <div>
                    <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
                      Stats
                    </h4>
                    <div className="p-3 bg-card rounded-lg border border-border space-y-2">
                      <div className="flex justify-between text-xs">
                        <span className="text-muted-foreground">Hours Saved</span>
                        <span className="font-bold">12.5h</span>
                      </div>
                      <div className="flex justify-between text-xs">
                        <span className="text-muted-foreground">Noise Removed</span>
                        <span className="font-bold">45m</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Main Content */}
              <div className="lg:col-span-9 p-6 md:p-8">
                <div className="flex items-center justify-between mb-8">
                  <div>
                    <h2 className="text-2xl font-bold">Green's Theorem & Vector Fields</h2>
                    <p className="text-muted-foreground">
                      Prof. Sarah Mitchell • Lecture 14 • 54 mins
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <Badge
                      variant="outline"
                      className="border-green-500/30 text-green-600 bg-green-500/5"
                    >
                      <CheckCircle2 className="w-3 h-3 mr-1" /> Processed
                    </Badge>
                    <Badge
                      variant="outline"
                      className="border-blue-500/30 text-blue-600 bg-blue-500/5"
                    >
                      <Brain className="w-3 h-3 mr-1" /> AI Insights Ready
                    </Badge>
                  </div>
                </div>

                {/* Audio Visualizer */}
                <div className="mb-8 p-4 bg-muted/20 rounded-xl border border-border">
                  <div className="flex items-center gap-4 mb-2">
                    <Button
                      size="icon"
                      variant="ghost"
                      className="h-8 w-8 rounded-full bg-primary text-primary-foreground hover:bg-primary/90"
                    >
                      <Play className="w-3 h-3 ml-0.5" />
                    </Button>
                    <WaveBars />
                    <span className="text-xs font-mono text-muted-foreground">14:20 / 54:00</span>
                  </div>
                  <div className="flex justify-between px-12">
                    <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-medium">
                      Prof. Mitchell
                    </span>
                    <span className="text-[10px] text-destructive/60 uppercase tracking-widest font-medium line-through">
                      Student Interruptions
                    </span>
                    <span className="text-[10px] text-muted-foreground uppercase tracking-widest font-medium">
                      Prof. Mitchell
                    </span>
                  </div>
                </div>

                <div className="grid md:grid-cols-2 gap-8">
                  {/* Transcript */}
                  <div className="space-y-4">
                    <h3 className="font-semibold flex items-center gap-2 text-sm uppercase tracking-wider text-muted-foreground">
                      <FileText className="w-4 h-4" /> Live Transcript
                    </h3>
                    <div className="space-y-4 text-sm leading-relaxed">
                      <div className="pl-4 border-l-2 border-primary">
                        <p className="font-semibold text-primary mb-1">Prof. Mitchell</p>
                        <p>
                          ...which brings us to the core concept. Green's Theorem relates a line
                          integral around a simple closed curve C to a double integral over the
                          plane region D bounded by C.
                        </p>
                      </div>
                      <div className="pl-4 border-l-2 border-destructive/20 bg-destructive/5 p-2 rounded-r-md opacity-60">
                        <div className="flex justify-between items-center mb-1">
                          <p className="font-semibold text-destructive text-xs">
                            Noise Detected (Filtered)
                          </p>
                          <Badge
                            variant="outline"
                            className="text-[10px] h-5 px-1 border-destructive/20 text-destructive"
                          >
                            Removed
                          </Badge>
                        </div>
                        <p className="text-muted-foreground italic text-xs">
                          Student: (coughing) ...can you repeat that last part about the curve?
                        </p>
                      </div>
                      <div className="pl-4 border-l-2 border-primary">
                        <p className="font-semibold text-primary mb-1">Prof. Mitchell</p>
                        <p>
                          This is crucial for flux calculations in engineering. Think of it as the
                          2D version of Stokes' Theorem.
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* AI Insights */}
                  <div className="space-y-4">
                    <h3 className="font-semibold flex items-center gap-2 text-sm uppercase tracking-wider text-muted-foreground">
                      <Sparkles className="w-4 h-4" /> AI Knowledge Graph
                    </h3>

                    <div className="bg-card border border-border rounded-lg p-4 shadow-sm">
                      <h4 className="font-semibold mb-2 text-sm">Key Concepts Detected</h4>
                      <div className="flex flex-wrap gap-2">
                        <Badge
                          variant="secondary"
                          className="hover:bg-primary/10 cursor-pointer transition-colors"
                        >
                          Green's Theorem
                        </Badge>
                        <Badge
                          variant="secondary"
                          className="hover:bg-primary/10 cursor-pointer transition-colors"
                        >
                          Line Integral
                        </Badge>
                        <Badge
                          variant="secondary"
                          className="hover:bg-primary/10 cursor-pointer transition-colors"
                        >
                          Flux
                        </Badge>
                        <Badge
                          variant="secondary"
                          className="hover:bg-primary/10 cursor-pointer transition-colors"
                        >
                          Stokes' Theorem
                        </Badge>
                      </div>
                    </div>

                    {isV2 && (
                      <div className="bg-linear-to-br from-purple-50 to-blue-50 dark:from-purple-900/10 dark:to-blue-900/10 border border-purple-200 dark:border-purple-800 rounded-lg p-4 shadow-sm relative overflow-hidden">
                        <div className="flex items-center gap-2 mb-3">
                          <Library className="w-4 h-4 text-purple-600" />
                          <h4 className="font-semibold text-sm text-purple-900 dark:text-purple-100">
                            Enrichment Material
                          </h4>
                        </div>
                        <ul className="space-y-3">
                          <li className="text-xs flex gap-2 items-start">
                            <span className="bg-white dark:bg-black/20 text-purple-600 rounded text-[10px] px-1 font-mono border border-purple-200">
                              PAPER
                            </span>
                            <span className="text-muted-foreground">
                              "Applications of Green's Theorem in Fluid Dynamics" - Journal of
                              Physics, 2018
                            </span>
                          </li>
                          <li className="text-xs flex gap-2 items-start">
                            <span className="bg-white dark:bg-black/20 text-blue-600 rounded text-[10px] px-1 font-mono border border-blue-200">
                              BOOK
                            </span>
                            <span className="text-muted-foreground">
                              Vector Calculus, Marsden & Tromba (Chapter 7)
                            </span>
                          </li>
                        </ul>
                      </div>
                    )}

                    <div className="bg-card border border-border rounded-lg p-4 shadow-sm">
                      <h4 className="font-semibold mb-2 text-sm">Action Items</h4>
                      <ul className="space-y-2">
                        <li className="flex items-center gap-2 text-xs text-muted-foreground">
                          <div className="w-4 h-4 rounded border border-primary/50 flex items-center justify-center"></div>
                          Review practice problems on Flux
                        </li>
                        <li className="flex items-center gap-2 text-xs text-muted-foreground">
                          <div className="w-4 h-4 rounded border border-primary/50 flex items-center justify-center"></div>
                          Prepare for quiz on Friday
                        </li>
                      </ul>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section id="features" className="py-24">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Why top students rely on LectureAI
            </h2>
            <p className="text-muted-foreground text-lg">
              We don't just record audio. We engineer a learning environment that extracts signal
              from noise.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon={<VolumeX className="w-6 h-6" />}
              title="Intelligent Noise Gate"
              description="Our proprietary algorithm identifies the frequency footprint of your professor and suppresses everything else—coughing, whispering, and HVAC noise."
            />
            <FeatureCard
              icon={<Users className="w-6 h-6" />}
              title="Student Voice Filtering"
              description="Irrelevant questions from the back of the hall? Gone. We detect and lower the volume of non-instructional voices automatically."
            />
            <FeatureCard
              icon={<Zap className="w-6 h-6" />}
              title="Instant Summarization"
              description="Walk out of class with a study guide. We generate bullet-point summaries, key term definitions, and homework lists instantly."
            />
            <FeatureCard
              icon={<Library className="w-6 h-6" />}
              title={isV2 ? 'Deep Knowledge Search' : 'Contextual Linking'}
              description={
                isV2
                  ? "We don't just transcribe. We search millions of academic papers to find sources that verify and expand upon what your professor is teaching."
                  : 'Tap any word in your transcript to get a definition, Wikipedia summary, or related concept map.'
              }
            />
            <FeatureCard
              icon={<Mic className="w-6 h-6" />}
              title="Studio-Quality Enhancement"
              description="Even if you're sitting in the back row, our audio enhancement makes it sound like the professor is speaking directly into your headphones."
            />
            <FeatureCard
              icon={<Brain className="w-6 h-6" />}
              title="Focus Mode"
              description="Playback lectures at 2x speed with 'Silence Skipping' to consume an hour-long lecture in 20 minutes without missing a word."
            />
          </div>
        </div>
      </section>

      {/* Comparison Section */}
      <section className="py-24 bg-muted/50 border-y border-border/50">
        <div className="container mx-auto px-6">
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div className="space-y-6">
              <h2 className="text-3xl font-bold">Stop studying the hard way.</h2>
              <p className="text-lg text-muted-foreground">
                Traditional voice memos are messy, unsearchable, and full of distractions. LectureAI
                transforms chaos into clarity.
              </p>

              <div className="space-y-4">
                <div className="flex items-start gap-4">
                  <div className="w-8 h-8 rounded-full bg-red-100 dark:bg-red-900/20 flex items-center justify-center text-red-600 mt-1">
                    <VolumeX className="w-4 h-4" />
                  </div>
                  <div>
                    <h4 className="font-semibold">Standard Voice Memos</h4>
                    <p className="text-sm text-muted-foreground">
                      Full of background noise, hard to hear the professor, impossible to search.
                    </p>
                  </div>
                </div>
                <div className="flex items-start gap-4">
                  <div className="w-8 h-8 rounded-full bg-green-100 dark:bg-green-900/20 flex items-center justify-center text-green-600 mt-1">
                    <CheckCircle2 className="w-4 h-4" />
                  </div>
                  <div>
                    <h4 className="font-semibold">LectureAI Processing</h4>
                    <p className="text-sm text-muted-foreground">
                      Crystal clear audio, fully searchable transcript, distraction-free.
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="relative">
              <div className="absolute inset-0 bg-linear-to-tr from-primary/20 to-purple-500/20 blur-3xl rounded-full" />
              <div className="relative bg-background border border-border rounded-2xl p-8 shadow-2xl">
                <div className="space-y-4">
                  <div className="flex items-center justify-between border-b border-border pb-4">
                    <span className="font-semibold">Comparison</span>
                  </div>
                  <div className="space-y-6">
                    <div>
                      <div className="flex justify-between text-sm mb-2">
                        <span className="text-muted-foreground">Voice Memo App</span>
                        <span className="text-red-500 text-xs font-bold">NOISY</span>
                      </div>
                      <div className="h-12 bg-muted rounded-md flex items-center justify-center overflow-hidden relative">
                        <CompareBars variant="noisy" />
                      </div>
                    </div>
                    <div>
                      <div className="flex justify-between text-sm mb-2">
                        <span className="text-muted-foreground">LectureAI</span>
                        <span className="text-green-500 text-xs font-bold">CLEAN</span>
                      </div>
                      <div className="h-12 bg-primary/5 border border-primary/20 rounded-md flex items-center justify-center overflow-hidden relative">
                        <CompareBars variant="clean" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section id="testimonials" className="py-24">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">Loved by students everywhere</h2>
            <p className="text-muted-foreground text-lg">
              See what your peers are saying about LectureAI.
            </p>
          </div>
          <div className="grid md:grid-cols-3 gap-8">
            {[
              {
                name: 'Sarah Jenkins',
                role: 'Pre-med Student',
                quote:
                  'LectureAI saved my biochemistry grade. The noise cancellation is magic—I can finally hear the professor clearly.',
              },
              {
                name: 'David Chen',
                role: 'Law Student',
                quote:
                  'The summaries are a lifesaver. I used to spend hours re-listening to lectures, now I just review the AI notes.',
              },
              {
                name: 'Emily Rodriguez',
                role: 'Computer Science',
                quote:
                  'Being able to search through the transcript for specific algorithms mentioned in class is a game changer.',
              },
            ].map((t, i) => (
              <div key={i} className="bg-card p-6 rounded-2xl border border-border shadow-sm">
                <div className="flex items-center gap-1 mb-4 text-primary">
                  {Array.from({
                    length: 5,
                  }).map((_, j) => (
                    <Sparkles key={j} className="w-4 h-4 fill-current" />
                  ))}
                </div>
                <p className="text-muted-foreground mb-6">"{t.quote}"</p>
                <div>
                  <p className="font-semibold">{t.name}</p>
                  <p className="text-xs text-muted-foreground">{t.role}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section id="pricing" className="py-24 bg-muted/30 border-y border-border/50">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16 max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">Simple, transparent pricing</h2>
            <p className="text-muted-foreground text-lg">
              Invest in your education for less than the price of a coffee per week.
            </p>
          </div>
          <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
            {/* Free Tier */}
            <div className="bg-card p-8 rounded-2xl border border-border shadow-sm relative">
              <h3 className="text-xl font-bold mb-2">Free</h3>
              <div className="text-3xl font-bold mb-6">
                $0<span className="text-base font-normal text-muted-foreground">/mo</span>
              </div>
              <ul className="space-y-3 mb-8 text-sm text-muted-foreground">
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> 5 hours recording/mo
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> Basic noise reduction
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> Automated transcripts
                </li>
              </ul>
              <Button variant="outline" className="w-full">
                Get Started
              </Button>
            </div>

            {/* Pro Tier */}
            <div className="bg-card p-8 rounded-2xl border-2 border-primary shadow-xl relative scale-105 z-10">
              <div className="absolute top-0 right-0 bg-primary text-primary-foreground text-xs font-bold px-3 py-1 rounded-bl-xl rounded-tr-lg">
                MOST POPULAR
              </div>
              <h3 className="text-xl font-bold mb-2">Pro</h3>
              <div className="text-3xl font-bold mb-6">
                $12<span className="text-base font-normal text-muted-foreground">/mo</span>
              </div>
              <ul className="space-y-3 mb-8 text-sm text-muted-foreground">
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> Unlimited recording
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> Advanced Audio Isolation 2.0
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> AI Summaries & Key Concepts
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> Search across all lectures
                </li>
              </ul>
              <Button className="w-full shadow-lg shadow-primary/20">Upgrade to Pro</Button>
            </div>

            {/* University Tier */}
            <div className="bg-card p-8 rounded-2xl border border-border shadow-sm relative">
              <h3 className="text-xl font-bold mb-2">University</h3>
              <div className="text-3xl font-bold mb-6">Custom</div>
              <ul className="space-y-3 mb-8 text-sm text-muted-foreground">
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> Site-wide license
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> LMS Integration (Canvas,
                  Blackboard)
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-primary" /> Admin dashboard
                </li>
              </ul>
              <Button variant="outline" className="w-full">
                Contact Sales
              </Button>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-32 relative overflow-hidden">
        <div className="absolute inset-0 bg-primary/5" />
        <div className="container mx-auto px-6 text-center relative z-10">
          <div className="max-w-3xl mx-auto space-y-8">
            <h2 className="text-4xl md:text-5xl font-extrabold tracking-tight">
              Join 50,000+ students achieving <br />
              <span className="text-primary">better grades with less stress.</span>
            </h2>
            <p className="text-xl text-muted-foreground">
              Stop worrying about taking notes. Start listening. Let LectureAI handle the rest.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center pt-4">
              <Button
                size="lg"
                className="h-14 px-10 text-lg rounded-full shadow-lg hover:shadow-xl transition-all"
              >
                Get Your First Month Free
              </Button>
              <Button
                size="lg"
                variant="outline"
                className="h-14 px-10 text-lg rounded-full bg-background"
              >
                View Pricing
              </Button>
            </div>
            <p className="text-sm text-muted-foreground pt-4">
              No credit card required for trial • Cancel anytime
            </p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 border-t border-border bg-muted/5">
        <div className="container mx-auto px-6">
          <div className="grid md:grid-cols-4 gap-8 mb-12">
            <div className="col-span-1 md:col-span-2">
              <div className="flex items-center gap-2 mb-4">
                <Brain className="w-5 h-5 text-primary" />
                <span className="font-bold text-xl">LectureAI</span>
              </div>
              <p className="text-muted-foreground max-w-xs">
                Empowering students with AI-driven learning tools. We believe education should be
                accessible, clear, and distraction-free.
              </p>
            </div>
            <div>
              <h4 className="font-semibold mb-4">Product</h4>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li>
                  <a href="#" className="hover:text-primary">
                    Features
                  </a>
                </li>
                <li>
                  <a href="#" className="hover:text-primary">
                    Pricing
                  </a>
                </li>
                <li>
                  <a href="#" className="hover:text-primary">
                    For Universities
                  </a>
                </li>
                <li>
                  <a href="#" className="hover:text-primary">
                    Success Stories
                  </a>
                </li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-4">Company</h4>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li>
                  <a href="#" className="hover:text-primary">
                    About Us
                  </a>
                </li>
                <li>
                  <a href="#" className="hover:text-primary">
                    Careers
                  </a>
                </li>
                <li>
                  <a href="#" className="hover:text-primary">
                    Blog
                  </a>
                </li>
                <li>
                  <a href="#" className="hover:text-primary">
                    Contact
                  </a>
                </li>
              </ul>
            </div>
          </div>
          <div className="pt-8 border-t border-border flex flex-col md:flex-row justify-between items-center gap-4">
            <p className="text-sm text-muted-foreground">
              © 2025 LectureAI Inc. All rights reserved.
            </p>
            <div className="flex gap-6 text-sm text-muted-foreground">
              <a href="#" className="hover:text-foreground">
                Privacy Policy
              </a>
              <a href="#" className="hover:text-foreground">
                Terms of Service
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: ReactNode
  title: string
  description: string
}) {
  return (
    <div className="bg-card p-8 rounded-2xl border border-border shadow-sm hover:shadow-md hover:border-primary/20 transition-all duration-300 group">
      <div className="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center text-primary mb-6 group-hover:scale-110 transition-transform">
        {icon}
      </div>
      <h3 className="text-xl font-bold mb-3">{title}</h3>
      <p className="text-muted-foreground leading-relaxed">{description}</p>
    </div>
  )
}
