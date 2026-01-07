run:
	wails build -m -nosyncgomod -devtools -tags webkit2_41 && ./build/bin/badger-gui

run-dev:
	wails dev -m -nosyncgomod -tags webkit2_41