import React from 'react';
import './style.scss';
import Lottie from 'lottie-react';

const TitleComponent = (props) => {
    const { headerTitle, typeTitle = 'header', headerDescription, style, img, finish, required } = props;

    return (
        <div className="title-container" style={style?.container}>
            {typeTitle === 'header' && (
                <div className={finish ? 'header-title-container-finish' : 'header-title-container'}>
                    {finish ? (
                        <Lottie style={style?.image} animationData={img} loop={true} />
                    ) : (
                        <img className="header-image" src={img} alt={img} style={style?.image}></img>
                    )}

                    <label className="header-title" style={style?.header}>
                        {headerTitle}
                    </label>
                </div>
            )}
            {typeTitle === 'sub-header' && (
                <p className="sub-header-title" style={style?.header}>
                    {required && <span>* </span>}
                    {headerTitle}
                </p>
            )}
            <p className="header-description" style={style?.description}>
                {headerDescription}
            </p>
        </div>
    );
};

export default TitleComponent;
