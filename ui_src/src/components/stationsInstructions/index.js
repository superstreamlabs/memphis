import Button from '../../components/button';
import React from 'react';
import './style.scss';

const StationsInstructions = (props) => {
    const { header, button, image, newStation } = props;
    const description = 'A station is a distributed unit that producers store data at and consumers consume data from.';

    return (
        <div id="e2e-create-station-btn" className={image ? 'stations-placeholder' : 'stations-placeholder add-more'}>
            {image ? <img src={image} className="stations-icon" alt="stationsImage"></img> : null}
            <div className="header-empty-stations">{header}</div>
            <p className="header-empty-description">{description}</p>
            <Button
                className="modal-btn"
                width="180px"
                height="37px"
                placeholder={button}
                colorType={image ? 'white' : 'purple'}
                radiusType="circle"
                border={image ? 'none' : 'gray-light'}
                backgroundColorType={image ? 'purple' : 'none'}
                fontSize="14px"
                fontWeight="bold"
                onClick={() => newStation()}
            />
        </div>
    );
};

export default StationsInstructions;
