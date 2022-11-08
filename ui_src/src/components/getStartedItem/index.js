import React, { useContext } from 'react';
import './style.scss';
import Button from '../button';
import { GetStartedStoreContext } from '../../domain/overview/getStarted';
import bgGetStarted from '../../assets/images/bgGetStarted.svg';
import bgGetStartedBottom from '../../assets/images/bgGetStartedBottom.svg';
import VideoPlayer from '../videoPlayer';

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
                            <div>
                                <VideoPlayer url={'https://www.youtube.com/watch?v=kp_l2qFokso'} />
                                <p className="video-description">Watch starting Demo video ✨</p>
                            </div>
                            <div>
                                <VideoPlayer url={'https://www.youtube.com/watch?v=kp_l2qFokso'} />
                                <p className="video-description">Explore the system in 3 min ⭐</p>
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
