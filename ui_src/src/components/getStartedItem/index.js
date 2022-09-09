import React, { useContext } from 'react';
import TitleComponent from '../titleComponent';
import './style.scss';
import Button from '../button';
import { GetStartedStoreContext } from '../../domain/overview/getStarted';

const GetStartedItem = (props) => {
    const { headerImage, headerTitle, headerDescription, style, children, onNext, onBack } = props;
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);

    return (
        <div className="get-started-wrapper">
            <div className="get-started-top">
                <div className={getStartedState?.currentStep === 5 ? 'get-started-header finish' : 'get-started-header'}>
                    <TitleComponent
                        img={headerImage}
                        headerTitle={headerTitle}
                        headerDescription={headerDescription}
                        style={style}
                        finish={getStartedState?.currentStep === 5}
                    ></TitleComponent>
                </div>
                <div className="get-started-body">{children}</div>
            </div>
            <div className="get-started-footer">
                {!getStartedState.isHiddenButton && (
                    <>
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
                                onClick={() => onBack()}
                                isLoading={getStartedState?.isLoading}
                            />
                        )}
                    </>
                )}
            </div>
        </div>
    );
};

export default GetStartedItem;
