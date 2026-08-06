import Cocoa
import FlutterMacOS

class MainFlutterWindow: NSWindow {
  override func awakeFromNib() {
    let flutterViewController = FlutterViewController()
    let windowFrame = self.frame
    self.contentViewController = flutterViewController
    self.setFrame(windowFrame, display: true)

    RegisterGeneratedPlugins(registry: flutterViewController)

    super.awakeFromNib()
  }

  // window_manager's `waitUntilReadyToShow` contract assumes the native
  // window stays invisible until the Dart side explicitly calls
  // `windowManager.show()`. Without this override, AppKit's default
  // "visible at launch" nib behavior orders the window onto screen with
  // whatever title bar style is active at that moment — before Dart has
  // had a chance to apply `titleBarStyle`/`windowButtonVisibility` — which
  // is why the custom title bar never appeared. `hiddenWindowAtLaunch()`
  // is window_manager's own NSWindow extension (see its official example
  // app's MainFlutterWindow.swift); it's a one-shot `setIsVisible(false)`.
  override public func order(_ place: NSWindow.OrderingMode, relativeTo otherWin: Int) {
    super.order(place, relativeTo: otherWin)
    hiddenWindowAtLaunch()
  }
}
