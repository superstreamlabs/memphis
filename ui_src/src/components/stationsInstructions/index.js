import Button from 'components/button';
import React from 'react';
import { isCloud } from 'services/valueConvertor';
import LockFeature from 'components/lockFeature';
import './style.scss';

const StationsInstructions = ({ upgrade, header, button, image, newStation, des }) => {
    const description = des || 'A station is a distributed storage for messages';

    return (
        <div className={image ? 'stations-placeholder' : 'stations-placeholder add-more'}>
            {image ? <img src={image} className="stations-icon" alt="stationsImage"></img> : null}
            <div className="header-empty-stations">{header}</div>
            <p className="header-empty-description">{description}</p>
            {button && (
                <Button
                    className="modal-btn"
                    width="180px"
                    height="37px"
                    // placeholder={button}
                    placeholder={
                        isCloud() && upgrade ? (
                            <span className="upgrade">
                                {button} <LockFeature />
                            </span>
                        ) : (
                            button
                        )
                    }
                    colorType={image ? 'white' : 'purple'}
                    radiusType="circle"
                    border={image ? 'none' : 'gray-light'}
                    backgroundColorType={image ? 'purple' : 'none'}
                    fontSize="14px"
                    fontWeight="bold"
                    onClick={() => newStation()}
                />
            )}
        </div>
    );
};

export default StationsInstructions;
