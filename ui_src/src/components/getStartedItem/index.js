import React, { useContext } from 'react';
import './style.scss';
import Button from '../button';
import { GetStartedStoreContext } from '../../domain/overview/getStarted';
import bgGetStarted from '../../assets/images/bgGetStarted.svg';
import bgGetStartedBottom from '../../assets/images/bgGetStartedBottom.svg';
import VideoPlayer from '../videoPlayer';
import blackBall from '../../assets/images/blackBall.svg';
import orangeBall from '../../assets/images/orangeBall.svg';
import pinkBall from '../../assets/images/pinkBall.svg';
import purpleBall from '../../assets/images/purpleBall.svg';
import { CONNECT_APP_VIDEO, CONNECT_CLI_VIDEO } from '../../config';

const GetStartedItem = (props) => {
    const { headerImage, headerTitle, headerDescription, style, children, onNext, onBack } = props;
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);

    return (
        <div className="get-started-wrapper">
            {getStartedState?.currentStep !== 5 && (
                <>
                    <img className="get-started-bg-img" src={bgGetStarted} alt="bgGetStarted" />
                    <div className="get-started-top">
                        <div className="get-started-top-header">
                            <img className="header-image" src={headerImage} alt={headerImage} />
                            <p className="header-title">{headerTitle}</p>
                            <div className="header-description">{headerDescription}</div>
                        </div>
                        <div className="get-started-body">{children}</div>
                    </div>
                </>
            )}
            {getStartedState?.currentStep === 5 && (
                <>
                    <img className="get-started-bg-img" src={bgGetStarted} alt="bgGetStarted" />
                    <img className="get-started-bg-img-bottom" src={bgGetStartedBottom} alt="bgGetStartedBottom"></img>
                    <div className="get-started-top">
                        <div className="video-container">
                            <div className="video-section">
                                <div className="video-section-black-ball">
                                    <img className="black-ball" src={blackBall} alt="black-ball"></img>
                                </div>
                                <img className="orange-ball" src={orangeBall} alt="orange-ball"></img>
                                <VideoPlayer url={CONNECT_APP_VIDEO} />
                                <p className="video-description">Connect your first app to Memphis ✨</p>
                            </div>
                            <div className="video-section">
                                <img className="pink-ball" src={pinkBall} alt="pink-ball"></img>
                                <img className="purple-ball" src={purpleBall} alt="purple-ball"></img>

                                <VideoPlayer url={CONNECT_CLI_VIDEO} />
                                <p className="video-description">How to install and connect Memphis.dev CLI ⭐</p>
                            </div>
                        </div>
                        <div className="get-started-top-header finish">
                            <p className="header-title">{headerTitle}</p>
                            <div className="header-description">{headerDescription}</div>
                        </div>
                        <div className="get-started-body-finish">{children}</div>
                    </div>
                </>
            )}
            {!getStartedState.isHiddenButton && getStartedState?.currentStep !== 5 && (
                <div className="get-started-footer">
                    <div id="e2e-getstarted-next-btn">
                        <Button
                            width={getStartedState?.currentStep === 5 ? '190px' : '129px'}
                            height="36px"
                            placeholder={getStartedState?.currentStep === 5 ? 'Launch Dashboard' : 'Next'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType={'purple'}
                            fontSize="16px"
                            fontWeight="bold"
                            htmlType="submit"
                            disabled={getStartedState?.nextDisable}
                            onClick={() => onNext()}
                            isLoading={getStartedState?.isLoading}
                        />
                    </div>
                    {getStartedState?.currentStep !== 1 && (
                        <Button
                            width={'129px'}
                            height="36px"
                            placeholder={'Back'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType={'black'}
                            fontSize="16px"
                            fontWeight="bold"
                            htmlType="submit"
                            disabled={getStartedState?.currentStep === 1}
                            onClick={() => onBack()}
                            isLoading={getStartedState?.isLoading}
                        />
                    )}
                </div>
            )}
        </div>
    );
};

export default GetStartedItem;
