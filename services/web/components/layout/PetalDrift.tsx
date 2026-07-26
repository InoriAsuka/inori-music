/**
 * PetalDrift — ambient sakura petals falling behind app content.
 *
 * Deliberately only six petals: enough to read as atmosphere, few enough that
 * the compositor cost stays negligible. Purely decorative, so it is hidden from
 * assistive tech and removed outright under prefers-reduced-motion (see the
 * `.petal` rule in globals.css — zeroing the duration would freeze petals
 * mid-screen instead of hiding them).
 */
const PETALS = [
  { left: "8%", size: 10, duration: 17, delay: 0, drift: "5vw", opacity: 0.35 },
  { left: "24%", size: 14, duration: 22, delay: 4, drift: "-3vw", opacity: 0.28 },
  { left: "43%", size: 9, duration: 19, delay: 9, drift: "7vw", opacity: 0.3 },
  { left: "61%", size: 13, duration: 25, delay: 2, drift: "-6vw", opacity: 0.24 },
  { left: "78%", size: 11, duration: 20, delay: 12, drift: "4vw", opacity: 0.32 },
  { left: "92%", size: 8, duration: 23, delay: 7, drift: "-5vw", opacity: 0.26 },
];

export function PetalDrift() {
  return (
    <div aria-hidden="true" className="pointer-events-none fixed inset-0 z-0 overflow-hidden">
      {PETALS.map((p) => (
        <span
          key={p.left}
          className="petal"
          style={
            {
              left: p.left,
              "--petal-size": `${p.size}px`,
              "--petal-duration": `${p.duration}s`,
              "--petal-delay": `${p.delay}s`,
              "--petal-drift": p.drift,
              "--petal-opacity": p.opacity,
            } as React.CSSProperties
          }
        />
      ))}
    </div>
  );
}
