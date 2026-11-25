import React, { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import MerchantLogo from '../components/MerchantLogo';
import Money from '../components/Money';

function SpendleStage() {
  const host = useRef(null);
  const [scale, setScale] = useState(1);
  const [desktop, setDesktop] = useState(false);
  useEffect(() => {
    const media = window.matchMedia('(min-width: 768px)');
    const resize = () => { setDesktop(media.matches); if (host.current) setScale(Math.min(1, host.current.clientWidth / (media.matches ? 1120 : 360))); };
    const observer = new ResizeObserver(resize);
    if (host.current) observer.observe(host.current);
    media.addEventListener('change', resize); resize();
    return () => { observer.disconnect(); media.removeEventListener('change', resize); };
  }, []);
  const width = desktop ? 1120 : 360, height = desktop ? 640 : 480;
  return <div ref={host} className="hero-stage-host" style={{ height: height * scale }}><div className="hero-artboard" style={{ width, height, transform: `translateX(-50%) scale(${scale})` }}>
    <div className="stage-orbit orbit-one"/><div className="stage-orbit orbit-two"/>
    <div className="stage-phone phone-left"><span className="phone-notch"/><small>OCTOBER · 2026</small><b>₹4,250</b><span>safe to spend</span><div className="phone-bars"><i/><i/><i/></div></div>
    <article className="stage-center"><div className="stage-card-top"><span>Spendle<span>.</span></span><i>SAMPLE</i></div><span className="stage-kicker">LEFT THIS MONTH</span><div className="stage-balance"><Money paise={425000}/></div><div className="stage-day">DAY 18 / 30</div><div className="stage-progress"><i/></div><div className="stage-card-bottom">UPI · CASH · CARD</div></article>
    <div className="stage-phone phone-right"><span className="phone-notch"/><small>YOUR BUDGET</small><b>Food & dining</b><span>₹3,420 of ₹6,000</span><div className="phone-bars"><i/><i/><i/></div></div>
    <span className="stage-chip chip-a"><MerchantLogo name="zomato"/><span>Zomato <b><Money paise={34000}/></b></span><i/></span><span className="stage-chip chip-b">Budget <b>57%</b><i/></span><span className="stage-chip chip-c">Bike goal <b>38%</b><i/></span>
    <span className="stage-dot"/>
  </div></div>;
}

export default function Hero() {
  return <section id="hero-section" className="landing-hero">
    <div className="hero-copy">
      <p className="hero-eyebrow"><i/> PERSONAL FINANCE · ZERO SPREADSHEETS</p>
      <h1>Know what's<br/><em>left.</em></h1>
      <p className="hero-subline">See what’s left this month. Track every UPI spend, set simple budgets, and keep your goals moving.</p>
      <div className="hero-actions"><Link to="/login?mode=signup" className="hero-start">Get started</Link><a href="#month-section" className="hero-how">See how a month works <span>↓</span></a></div>
    </div>
    <SpendleStage/>
  </section>;
}
