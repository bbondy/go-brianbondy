package data

// CareerProfile is the single source of truth for the concise resume and the
// full career archive. Role metadata is shared; each page selects the level of
// detail appropriate to its audience.
type CareerProfile struct {
	Name               string
	Headline           string
	ExecutiveSummary   string
	ScaleProof         CareerScaleProof
	Links              []CareerLink
	Roles              []CareerRole
	TechnicalDepth     []CareerSkillGroup
	Timeline           []CareerTimelineEntry
	TechnicalInventory []CareerSkillGroup
	Projects           []CareerProject
	Education          CareerEducation
	Interests          []string
	ArchiveBackground  []string
}

type CareerScaleProof struct {
	Value       string
	Label       string
	SourceLabel string
	SourceURL   string
}

type CareerLink struct {
	Label string
	URL   string
}

type CareerRole struct {
	Company           string
	CompanyURL        string
	Title             string
	Location          string
	Dates             string
	Summary           string
	ResumeHighlights  []CareerHighlight
	ArchiveHighlights []CareerHighlight
	ShowOnResume      bool
	OpenInArchive     bool
}

type CareerHighlight struct {
	Text        string
	SourceLabel string
	SourceURL   string
}

type CareerSkillGroup struct {
	Name  string
	Items string
}

type CareerTimelineEntry struct {
	Era          string
	Title        string
	Description  string
	Technologies string
}

type CareerProject struct {
	Name        string
	Description string
	URL         string
}

type CareerEducation struct {
	School      string
	Degree      string
	Location    string
	Dates       string
	Coursework  string
	Development string
}

// CareerProfileData returns the content used by both career pages.
func CareerProfileData() CareerProfile {
	return CareerProfile{
		Name:             "Brian R. Bondy",
		Headline:         "Co-founder & CTO, Brave Software",
		ExecutiveSummary: "Co-founder and CTO of Brave Software and a member of its board. Since 2015, I have helped build Brave from its first browser prototypes into a privacy-focused browser and product company serving more than 120 million monthly active users. My work combines company and engineering leadership with hands-on architecture and development across browsers, security, operating systems, networking, privacy, AI, and web applications. Before Brave, I built Firefox platform technology at Mozilla, worked as a development lead at Khan Academy, and co-founded software businesses.",
		// Brave's official About page is the source for this deliberately rounded
		// threshold. It reported 120+ million MAU when reviewed in August 2026.
		ScaleProof: CareerScaleProof{
			Value:       "120M+",
			Label:       "monthly active Brave users worldwide",
			SourceLabel: "Brave company data",
			SourceURL:   "https://brave.com/about/",
		},
		Links: []CareerLink{
			{Label: "Email", URL: "mailto:bbondy@gmail.com"},
			{Label: "GitHub", URL: "https://github.com/bbondy"},
			{Label: "LinkedIn", URL: "https://www.linkedin.com/in/bbondy"},
			{Label: "Stack Overflow", URL: "https://stackoverflow.com/users/3153/brian-r-bondy"},
			{Label: "X / Twitter", URL: "https://twitter.com/brianbondy"},
			{Label: "brianbondy.com", URL: "https://brianbondy.com"},
		},
		Roles: []CareerRole{
			{
				Company:       "Brave Software",
				CompanyURL:    "https://brave.com",
				Title:         "Co-founder and CTO",
				Location:      "Remote",
				Dates:         "May 2015–Present",
				Summary:       "Co-founded Brave and helped take it from an initial browser concept to a global privacy product company, while remaining deeply involved in architecture and implementation.",
				ShowOnResume:  true,
				OpenInArchive: true,
				ResumeHighlights: []CareerHighlight{
					{Text: "Co-founded the company with Brendan Eich and helped define and build the original product: privacy-first browsers for desktop, iOS, and Android, with blocking and privacy-preserving advertising built in."},
					{Text: "Led architecture and hands-on development through three major desktop foundations: the early Gecko/Graphene prototype, the shipped Muon/Electron implementation, and Brave Core, a maintainable Chromium fork that improved extension support, upgrade velocity, reliability, retention, and growth."},
					{Text: "Built and led engineering teams as the company scaled; managed as many as 60 employees and contractors, approximately one-third of the company, while recruiting and mentoring engineers."},
					{Text: "Led architecture and development across the browser, Brave Wallet, and Brave Leo. The broader platform grew to include Shields, native ad blocking, Tor windows, WebTorrent, IPFS, component updates, privacy-preserving ads, and desktop/mobile code sharing."},
					{Text: "Remain hands-on in the codebase across product architecture, technical investigations, features, fixes, developer infrastructure, and autonomous coding-agent tooling."},
				},
				ArchiveHighlights: []CareerHighlight{
					{Text: "During four months of pre-founding discussions, helped turn the initial pitch into a concrete technical plan covering browser architecture, installers, updates, signing, cost, schedule, and delivery.", SourceLabel: "Founding history", SourceURL: "/blog/174/the-road-to-brave-1-0"},
					{Text: "Started the first desktop codebase in May 2015 on Mozilla Graphene and Gecko. The team built a React/Redux browser UI from scratch, then stopped the unreleased prototype after 637 commits when missing platform support made the approach impractical."},
					{Text: "Moved the desktop browser to Muon, a security-hardened Electron fork with sandboxing enabled and Node removed from renderer processes. The team ported and shipped a beta-quality browser in one month and five days, then took Brave out of stealth on January 20, 2016."},
					{Text: "Helped navigate repeated mobile architecture decisions: an early React Native iOS attempt, a Firefox iOS–based release with the APIs needed for Shields, the Link Bubble acquisition and Android launch, and later Chromium/WebKit foundations."},
					{Text: "Led the transition from Muon to Brave Core, structured as a maintainable Chromium fork. It replaced six-week, two-engineer Chromium upgrades, enabled broad extension compatibility, increased code sharing with Android, and materially improved browser retention and growth."},
					{Text: "Led architecture and development of major products and capabilities including Brave Browser, Brave Wallet, and the privacy-preserving Leo AI assistant, while continuing to contribute smaller enhancements and fixes."},
					{Text: "The platform added Brave Shields, a high-performance ad-blocking engine, privacy-preserving ads and Rewards, Tor windows, WebTorrent, IPFS with an embedded node, and a native multi-chain wallet. These were company and team efforts, not individual projects."},
					{Text: "Built and mentored engineering teams and managed as many as 60 employees and contractors, roughly one-third of the company. Brave had grown from two people to more than 100 teammates by the 1.0 launch in 2019."},
					{Text: "Remain hands-on in Brave code and developer tooling, including browser work, the Go component-update service, and Brave Dev Bot, a Ralph Wiggum loop-inspired coding agent that implements stories, opens pull requests, responds to review, and manages CI loops."},
				},
			},
			{
				Company:    "Khan Academy",
				CompanyURL: "https://www.khanacademy.org",
				Title:      "Developer Lead",
				Location:   "Remote",
				// The 2015 retrospective says "Today is my last day," but its
				// manifest date is not sufficient evidence for an exact end month.
				Dates:         "April 2014–2015",
				Summary:       "Built learning environments and platform capabilities across Khan Academy's computer-science product, including a fully client-side SQL environment and the HTML/CSS environment.",
				ShowOnResume:  true,
				OpenInArchive: true,
				ResumeHighlights: []CareerHighlight{
					{Text: "Created a fully client-side environment for learning SQL using SQLite compiled to asm.js with Emscripten and sql.js; a complete Khan Academy SQL course was built around the environment."},
					{Text: "Helped generalize the computer-science platform beyond JavaScript, co-authored its HTML/CSS environment, and led peer project evaluations and LearnStorm, a three-month Bay Area math competition."},
					{Text: "Designed and implemented a Gecko-based end-to-end testing system used during deployments, and independently built the Khan Academy Firefox OS app as a side project."},
				},
				ArchiveHighlights: []CareerHighlight{
					{Text: "Created the SQL learning environment during hack week using SQLite, Emscripten, asm.js, and sql.js. It recreated database state and results on every edit while running entirely in the browser.", SourceLabel: "SQL implementation", SourceURL: "/blog/168/sql-on-khan-academy-enabled-by-sqlite-sqljs-asmjs-and-emscripten"},
					{Text: "Helped generalize the computer-science environment so it was no longer JavaScript-specific, and co-authored the HTML/CSS learning environment with John Resig, Pamela Fox, and Alex Rodrigues."},
					{Text: "Led development of peer evaluations for student coding projects and, with Colin Fuller, LearnStorm, a Google-funded three-month Bay Area math competition."},
					{Text: "Built an autosuggest code-editor feature as a three-day hackathon project and implemented a Gecko-based automated end-to-end testing framework in under two weeks; it was used during each deploy."},
					{Text: "Was sole primary developer of the Khan Academy Firefox OS app. It supported offline content, search, account progress, localization, low-memory devices, and the full video/article library; the original release received a five-out-of-five rating."},
					{Text: "Served as developer lead for a user-satisfaction team and made approximately 700 commits spanning product work, SEO, discussions, Google App Engine integration, privacy, fixes, and enhancements.", SourceLabel: "Khan Academy retrospective", SourceURL: "/blog/170/thank-you-khan-academy-switching-careers"},
				},
			},
			{
				Company:       "Mozilla",
				CompanyURL:    "https://www.mozilla.org",
				Title:         "Senior Firefox Platform Engineer",
				Location:      "Remote",
				Dates:         "July 2011–April 2014",
				Summary:       "Worked across Firefox platform code, with particular depth in Windows integration, browser updates, graphics, performance, and the Windows 8 browser.",
				ShowOnResume:  true,
				OpenInArchive: true,
				ResumeHighlights: []CareerHighlight{
					{Text: "Led development of the Mozilla Maintenance Service for secure, silent Firefox updates on Windows and delivered updater security improvements across service and non-service paths."},
					{Text: "Developed the Windows 8 Metro Firefox browser across initial platform work, graphics integration, file picking, contracts, browser registration, installation, updates, and dozens of supporting changes."},
					{Text: "Became a module peer for Firefox, Toolkit, ImageLib, and Widget:Win32; contributed hundreds of patches across C++, JavaScript, and Python, and reviewed and mentored other contributors."},
				},
				ArchiveHighlights: []CareerHighlight{
					{Text: "Joined as a three-month contractor on July 6, 2011 and converted to a full-time employee a few months later."},
					{Text: "Led the Windows UAC/silent-update project and Mozilla Maintenance Service, including security enhancements to update flows. The service let Firefox update without repeated elevation prompts while preserving user control.", SourceLabel: "Silent-update design", SourceURL: "/blog/125/mozilla-firefox-and-silent-updates"},
					{Text: "Built substantial parts of Firefox for the Windows 8 Metro environment: initial groundwork, graphics integration, file picker, Metro contracts, rendering support, browser registration, installer and updater behavior, and dozens of related tasks."},
					{Text: "Worked on Windows platform integration and Snappy startup performance, plus crash fixes, memory-leak fixes, intermittent-test failures, telemetry, and general Firefox enhancements."},
					{Text: "Added BMP and ICO encoders, Vista PNG-in-ICO support, refactored the ICO decoder to reuse the BMP decoder, and added hundreds of reference tests."},
					{Text: "In the first year alone, landed more than 215 patches across more than 170 resolved bugs and filed 200 bugs. Later resume versions summarize the full tenure as hundreds of contributions."},
					{Text: "Served as a module peer for Firefox, Toolkit, ImageLib, and Widget:Win32, reviewed patches, mentored bugs, and joined Mozilla's Coding Stewards group.", SourceLabel: "First-year retrospective", SourceURL: "/blog/144/retrospection-my-first-year-at-mozilla"},
					{Text: "Created Code Firefox and a Mozilla development cheat sheet to help new contributors navigate the codebase and build process."},
				},
			},
			{
				Company:  "VisionWorks Solutions Inc.",
				Title:    "Co-founder and Software Developer",
				Location: "Windsor, Ontario, Canada",
				// TODO: The prior resume says January 2004, while a 2011 blog
				// post says VisionWorks was founded in 2003. Preserve the resume
				// date until the discrepancy can be verified.
				Dates:         "January 2004–July 2011",
				Summary:       "Co-founded and built commercial Windows products in backup, storage, encryption, networking, and security; ROBOBAK and the VisionWorks product portfolio were acquired in 2011.",
				ShowOnResume:  true,
				OpenInArchive: true,
				ResumeHighlights: []CareerHighlight{
					{Text: "Built ROBOBAK enterprise backup software, several hundred thousand lines of code spanning deduplication, journaling, indexing, archiving, replication, and agentless backup, before its acquisition by KineticD."},
					{Text: "Created Vision Backup, a 130,000+ line C++ product that sold more than 3,000 copies, and File Access Manager, a filesystem filter driver with more than 50,000 outstanding licenses."},
					{Text: "Managed developers and custom software projects while building COM SDKs, encrypted-drive technology, Windows drivers, anti-spam/phishing network filters, and backup infrastructure."},
				},
				ArchiveHighlights: []CareerHighlight{
					{Text: "Co-founded VisionWorks Solutions and the ROBOBAK company. ROBOBAK delivered enterprise remote-office/branch-office backup with deduplication, journaling, indexing, archiving, agentless backup, and replication across several hundred thousand lines of code."},
					{Text: "Vision Backup comprised more than 130,000 lines of C++ and sold more than 3,000 copies. It supported LAN and FTP backup, schedules, compression, email notifications, incremental backups, online updates and activation, tape and optical media, and a COM plug-in system."},
					{Text: "Built File Access Manager, a filesystem filter driver that enabled backup access to exclusively opened and locked files; the prior resume records more than 50,000 outstanding licenses and licensing by several backup vendors."},
					{Text: "Managed and developed custom software projects, managed and maintained an encrypted-drive application, worked on TDI-filter-based spam and phishing protection, and created COM SDKs for multiple products."},
					{Text: "The ROBOBAK business and VisionWorks product portfolio were acquired by KineticD in 2011.", SourceLabel: "Acquisition note", SourceURL: "/blog/113/kineticd-acquires-robobak"},
				},
			},
			{
				Company:  "ALT Software",
				Title:    "Device Driver Developer",
				Location: "Waterloo, Ontario, Canada",
				// TODO: Verify whether this role continued concurrently after
				// VisionWorks began in January 2004; the prior resume lists both.
				Dates:         "August 2003–December 2004",
				Summary:       "Developed a device driver for a security video-capture camera and debugged several other projects.",
				OpenInArchive: false,
				ArchiveHighlights: []CareerHighlight{
					{Text: "Developed a device driver for a security video-capture camera and debugged several other projects."},
				},
			},
			{
				Company:       "Army Simulation Centre",
				Title:         "Linux Software Developer",
				Location:      "Kingston, Ontario, Canada",
				Dates:         "August 2002–August 2003",
				Summary:       "Built Linux/Qt synchronization and network-analysis software for military simulation systems, beginning as a co-op and continuing part-time during school.",
				OpenInArchive: false,
				ArchiveHighlights: []CareerHighlight{
					{Text: "Started during a co-op term and continued part-time during school in 2003."},
					{Text: "Created a briefcase application that synchronized one or more remote directories."},
					{Text: "Built an application integrated with simulation software that sniffed network packets and displayed complex results using a custom Qt graphing widget."},
					{Text: "Created additional C++/Qt/File Alteration Monitor utilities, packaged the software as an RPM, and wrote project and user documentation."},
				},
			},
			{
				Company:       "Corel Corporation",
				Title:         "Researcher and Software Developer",
				Location:      "Ottawa, Ontario, Canada",
				Dates:         "August 2001–August 2002",
				Summary:       "Worked on Paradox 10, early .NET and web-service technology, internal software distribution, and technology research.",
				OpenInArchive: false,
				ArchiveHighlights: []CareerHighlight{
					{Text: "Rewrote printing functionality in Paradox 10 and contributed to its first and second service packs."},
					{Text: "Worked on .NET-related projects, including XSLT transformation of SOAP documents."},
					{Text: "Created a generic website-update notification system used company-wide."},
					{Text: "Evaluated emerging technologies for possible use at Corel and developed new company ideas."},
				},
			},
		},
		TechnicalDepth: []CareerSkillGroup{
			{Name: "Languages", Items: "C++, Rust, Go, JavaScript, Python, C, C#, SQL"},
			{Name: "Browser & platform architecture", Items: "Chromium, Gecko, WebKit, desktop and mobile browsers, multi-process sandboxing, extensions, installers, secure updates, OS integration"},
			{Name: "Systems, security & networking", Items: "Privacy protections, native blocking, Windows drivers and services, cryptography, storage and backup, protocol development including TCP/UDP, HTTP(S), and FTP/FTPS/SFTP"},
			{Name: "AI & agents", Items: "Browser-integrated AI, privacy-preserving assistants, autonomous coding agents, PyTorch, TensorFlow, Keras, scikit-learn"},
			{Name: "Web & application architecture", Items: "React/Redux, Node, FastAPI, Google App Engine, REST, SQL/SQLite, distributed services"},
			{Name: "Decentralized systems & wallets", Items: "Multi-chain wallet architecture, self-custody key management, peer-to-peer protocols, IPFS, Tor, WebTorrent, privacy-preserving ads and rewards"},
		},
		Timeline: []CareerTimelineEntry{
			{Era: "1990s–2005", Title: "Programming and computer-science foundations", Description: "Began programming well before university, then developed breadth across algorithms, compilers, operating systems, graphics, networking, AI, databases, and concurrent systems at Waterloo.", Technologies: "C, C++, Java, Scheme, Haskell, Matlab, MIPS assembly, Unix, Linux, OpenGL, Qt, UDP"},
			{Era: "2001–2011", Title: "Windows systems and commercial products", Description: "Moved from product and driver work at Corel, the Army Simulation Centre, and ALT Software into co-founding software companies and shipping backup, storage, encryption, and networking products.", Technologies: "C++, Win32, COM, MFC, Windows services and drivers, TDI, filesystems, cryptography, FTP/SFTP/HTTP, SQL Server, tape and optical media"},
			{Era: "2011–2014", Title: "Firefox platform engineering", Description: "Worked in a large open-source browser codebase on secure updates, Windows integration, graphics and image codecs, performance, telemetry, and a new Windows 8 browser front end.", Technologies: "Gecko, C++, JavaScript, Python, XUL, XBL, Win32, Mercurial, GDB"},
			{Era: "2014–2015", Title: "Interactive learning platforms", Description: "Built browser-based programming environments, developer tooling, testing infrastructure, student evaluation systems, and cross-platform learning applications at Khan Academy.", Technologies: "JavaScript, Python, React, Backbone, SQLite, Emscripten, asm.js, sql.js, Google App Engine, Gecko, Firefox OS"},
			{Era: "2015–Present", Title: "Brave browser and privacy platform", Description: "Co-founded Brave and led through multiple browser architectures, a cross-platform Chromium foundation, privacy and security systems, Wallet, integrated AI, developer infrastructure, and growth to global scale.", Technologies: "Chromium, Gecko, WebKit, C++, Rust, Go, JavaScript, React/Redux, Node, IPFS, Tor, WebTorrent, cryptography, AI/LLMs"},
			{Era: "Recent", Title: "AI and autonomous development", Description: "Work now includes Brave Leo, machine-learning tools, autonomous coding agents, AI safety tooling, and compression experiments.", Technologies: "LLMs, agents, Python, PyTorch, TensorFlow, Keras, Rust, Go"},
		},
		TechnicalInventory: []CareerSkillGroup{
			{Name: "Languages and APIs", Items: "C; expert-level C++ including templates, STL, Boost, COM, MFC, Win32 API, OpenGL, and Qt; JavaScript for Node, clients, and applications; Python with TensorFlow, Keras, PyTorch, scikit-learn, Pandas, NumPy, and Django; C# with WinForms, WPF, Silverlight, WCF, LINQ, Entity Framework, remoting, generics, nullable types, and extension methods; F#, SQL, Matlab, Haskell, and Scheme."},
			{Name: "Server-side development", Items: "Node, IIS 5/6/7 and IIS extensions, Apache, CGI, FastAPI, Google App Engine runtimes, Flask, webapp, datastore, transactions, GQL, memcache, administration console, bulk data loaders/exporters, remote API, Django, ASP, ASP.NET, ASP.NET MVC 3 with ASPX and Razor, PHP, SOAP, REST/ROA, Jade, and Jinja2."},
			{Name: "Front-end development", Items: "DOM, CSS, Stylus, JavaScript, jQuery, jQuery UI, Backbone, React/Redux, debugging, AJAX, JSON, HTML5 canvas/video/local storage/web workers/offline applications/geolocation/forms/microdata, XML with XPath/XPointer/XSLT/SAX/DTDs, SVG, and MathML."},
			{Name: "Operating systems", Items: "macOS, Windows client and server systems including PE internals, Android, Unix, Solaris, Linux, BSD, iOS, and Windows Phone."},
			{Name: "Source control", Items: "Git, Mercurial, Subversion, CVS, Microsoft SourceSafe, and database IDEs."},
			{Name: "IDEs and debuggers", Items: "Vim, Cursor, Xcode, Microsoft Visual Studio versions 5 through 11 (including 2005, 2008, and 2010), Borland C++ 4.2, and GDB."},
			{Name: "Networking", Items: "Low-level socket programming; UDP and TCP; deep protocol knowledge including HTTP, HTTPS, FTP, FTPS, and POP; extensive packet analysis with Ethereal/Wireshark; client and server protocol implementations."},
			{Name: "Data stores", Items: "S3, Microsoft SQL Server 2000/2005/2008 including partitioning and index design, SQLite, DB2, MySQL, PostgreSQL, Oracle, Corel Paradox, and Microsoft Access."},
			{Name: "Systems and libraries", Items: "zlib, libxml2, image analysis, cryptography, steganography, LZW and Huffman compression, tape drives, CD-burning libraries, Windows LAN programming, NSIS installers, static and shared libraries (DLL/.so), semaphores, Windows multithreading, pthreads, shared memory, FreeType, and Video4Linux."},
		},
		Projects: []CareerProject{
			{Name: "Brave Browser", Description: "Privacy-focused browser for Windows, macOS, Linux, Android, and iOS with built-in blocking and privacy protections.", URL: "https://brave.com"},
			{Name: "Brave Wallet", Description: "Secure, native, multi-chain wallet built directly into Brave rather than delivered as an extension.", URL: "https://brave.com/wallet/"},
			{Name: "Brave Leo", Description: "AI assistant built into the browser for question answering, summarization, and other workflows with privacy protections.", URL: "https://brave.com/leo/"},
			{Name: "Brave Dev Bot", Description: "Ralph Wiggum loop-inspired coding agent for Brave projects that reads product requirements, implements user stories, runs tests, creates pull requests, handles review feedback, and manages CI in iterative loops.", URL: "https://github.com/brave-experiments/brave-dev-bot"},
			{Name: "IPFS in Brave", Description: "Deep browser integration of IPFS with CID-aware origin boundaries, public gateway support, an optional managed local node, component updates, and privacy-conscious lifecycle behavior.", URL: "/blog/177/ipfs-support-in-brave"},
			{Name: "Brave component update service", Description: "Go implementation of a Chromium-compatible component update server, built by observing Chromium's extension/component update protocol and flow.", URL: "https://github.com/brave/go-update"},
			{Name: "Code Firefox", Description: "A self-funded platform of short videos and exercises that explained how to become a Mozilla contributor. It received more than 100,000 unique visits and 30,000 full video views and helped hundreds of contributors and employees ramp up.", URL: "/blog/173/shutting-down-code-firefox"},
			{Name: "Codecheck.JS", Description: "Parses JavaScript with Acorn into an abstract syntax tree and checks the structure against assertions for use in online exercise frameworks.", URL: "https://github.com/bbondy/codecheckjs"},
			{Name: "Internet Library", Description: "Large C++ library built from scratch with FTP, FTPS, SFTP, SMTP, SMTP over SSL, POP3, HTTP/HTTPS clients and servers, cookies, proxies, HTML parsing, TCP, and UDP."},
			{Name: "Vision Backup", Description: "More than 130,000 lines of C++ for backup to disks, LAN systems, FTP-family servers, optical media, and tape, with a COM plug-in architecture; more than 3,000 copies sold."},
			{Name: "File Access Manager", Description: "Filesystem filter driver that allowed backup access to exclusively opened and locked files. The prior resume records more than 50,000 outstanding licenses and licensing by several backup companies."},
			{Name: "Cryptex", Description: "Virtual encrypted hard drive implemented as an NTFS driver. The vault appeared as a drive while unlocked and disappeared when locked."},
			{Name: "Firefox extensions", Description: "Extensions built with XUL, XBL, JavaScript, and XPCOM."},
			{Name: "Virtual disk drive", Description: "Presented remote data to Windows as a disk drive, with operations sent to an off-site IIS extension."},
			{Name: "Spam and phishing filter", Description: "TDI network filter for inspecting and filtering traffic, with Outlook and Outlook Express plug-ins."},
			{Name: "NullShare", Description: "Open-source peer-to-peer file-sharing application in C++ based on the Gnutella protocol."},
			{Name: "Pyroflow MSN", Description: "Complete multi-platform MSN Messenger client implementation built from scratch in C++ and Qt."},
			{Name: "Stego Flow", Description: "Image library and utilities for opening, inspecting, manipulating, and saving images, including steganographic encoding and extraction.", URL: "https://github.com/bbondy/stego-flow"},
			{Name: "Pyroflow Archiving", Description: "Cross-platform alternative to tar with built-in compression; it became the archive core of Vision Backup Enterprise."},
			{Name: "Data structures and compression", Description: "Template-based data structures and algorithms, including adaptive Huffman compression."},
		},
		Education: CareerEducation{
			School:      "University of Waterloo",
			Degree:      "Honours Bachelor of Mathematics in Computer Science",
			Location:    "Waterloo, Ontario, Canada",
			Dates:       "August 2000–April 2005",
			Coursework:  "Computer Graphics, Networking, Artificial Intelligence, Operating Systems, Algorithms, Concurrent Programming and Control Structures, Theory of Computation, Data Structures and Data Management, Number Theory, Mathematics of Investment, Statistics and Probability Theory, Combinatorics and Optimization, Logic, Calculus, and Classical and Linear Algebra.",
			Development: "Khan Academy learner from 2010 to the present as ongoing professional development.",
		},
		Interests: []string{
			"Ultrarunning, including completed events up to 250 miles",
			"Piano and guitar",
			"Reading and lifelong learning",
			"Strength training and martial arts; black belt (Shodan), Meibukan Goju-Ryu karate",
		},
		ArchiveBackground: []string{
			"Actively programming for more than 30 years.",
			"Bilingual in English and French.",
			"Extensive management, mentoring, and conflict-resolution experience.",
			"Microsoft MVP for Visual C++, July 2010–July 2011.",
			"Additional software roles and engagements include Evernote, Spoon.net, Telrex, Bluecherry, KineticD, RadioTime, and myBox.",
		},
	}
}
