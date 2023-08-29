// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { makeStyles, withStyles } from '@material-ui/core/styles';
import DialogContent from '@material-ui/core/DialogContent';
import Dialog from '@material-ui/core/Dialog';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import React, { useState } from 'react';

import FunctionsOverview from '../../components/functionsOverview';
import { getFontColor } from '../../utils/styleTemplates';
import { ReactComponent as SeperatorIcon } from '../../assets/images/seperator.svg';
import ConnectToHub from '../../components/connectToHub';
import { ReactComponent as ConnectIcon } from '../../assets/images/connect.svg';
import Button from '../../components/button';
import { ReactComponent as CloseIcon } from '../../assets/images/close.svg';
import FunctionDetails from './functionDetails';
import FunctionsList from './functionsList';

const AntTabs = withStyles({
    root: {
        paddingLeft: '30px'
    },
    indicator: {
        backgroundColor: getFontColor('black')
    }
})(Tabs);

const functions = [
    {
        _id: 1,
        name: 'sveta sveta sveta sveta sveta sveta sveta sveta',
        type: 'blabl'
    },
    {
        _id: 2,
        name: 'sveta2',
        type: 'blabl'
    },
    {
        _id: 3,
        name: 'sveta3',
        type: 'blabl'
    }
];

const AntTab = withStyles((theme) => ({
    root: {
        textTransform: 'none',
        fontSize: '16px',
        minWidth: 12,
        maxWidth: 100,
        fontWeight: theme.typography.fontWeightBold,
        marginRight: theme.spacing(3),
        fontFamily: ['Inter'].join(','),
        '&:hover': {
            color: getFontColor('navy'),
            opacity: 1
        },
        '&$selected': {
            color: getFontColor('navy'),
            fontWeight: theme.typography.fontWeightBold
        },
        '&:focus': {
            color: getFontColor('navy')
        }
    },
    selected: {}
}))((props) => <Tab disableRipple {...props} />);

const useStyles = makeStyles((theme) => ({
    root: {
        flexGrow: 1
    },
    dialogPaper: {
        height: '100vh',
        width: '100vw',
        borderRadius: '10px',
        minWidth: '1400px',
        overflowX: 'hidden',
        overflowY: 'scroll',
        position: 'relative'
    },
    dialogContent: {
        width: '100%',
        height: '100%',
        padding: '0px'
    }
}));

function HubMarketplace(props) {
    const classes = useStyles();
    const [value, setValue] = useState(0);
    const [chosenFunction, setChosenFunction] = useState(null);
    const [signedToHub, signInToHub] = useState(false); //placeholder - will be received from state
    const [isOpenSignIn, flipIsOpenSignIn] = useState(false);

    const handleChangeMenuItem = (_, newValue) => {
        setValue(newValue);
    };

    const handleCloseModal = () => {
        props.closeModal(false);
    };

    return (
        <Dialog
            open={props.open}
            onClose={(_, reson) => {
                if (reson === 'backdropClick') handleCloseModal();
                // { props.clickOutside() }
            }}
            classes={{ paper: classes.dialogPaper }}
        >
            <ConnectToHub open={isOpenSignIn} closeModeal={(e) => flipIsOpenSignIn(e)} />
            <div className="functions-modal-header">
                <div>
                    <label className="visit-hub">Visit hub</label>
                    <SeperatorIcon width={20} height={20} className="seperator" alt="seperator" />
                    {signedToHub ? (
                        <>
                            <ConnectIcon width={20} height={20} className="pointer" alt="connect to hub" />
                            <label className="sign-in-hub connected">Connected to hub</label>
                        </>
                    ) : (
                        <label className="sign-in-hub" onClick={() => flipIsOpenSignIn(true)}>
                            Sign in to hub
                        </label>
                    )}
                    <CloseIcon alt="close" width={12} height={12} style={{ cursor: 'pointer' }} onClick={handleCloseModal} />
                </div>
            </div>

            <DialogContent className={classes.dialogContent}>
                <AntTabs value={value} onChange={handleChangeMenuItem}>
                    <AntTab label="Private" />
                    <AntTab label="Public" />
                </AntTabs>
                <div className="functions-container">
                    <div className="function-list-section">
                        <FunctionsList chooseFunction={(func) => setChosenFunction(func)} />
                    </div>
                    <div className="function-details-section">
                        {' '}
                        <FunctionDetails chosenFunction={chosenFunction} />
                    </div>
                    {/* {value === 0 && <h1>Private</h1>}
                    {value === 1 && <h1>Public</h1>} */}
                </div>
                <div className="chosen-func-container">
                    <FunctionsOverview functions={functions} horizontal={true} editable={true} />
                </div>
                <div className="func-btn-footer">
                    <Button
                        className="modal-btn"
                        width="90px"
                        height="32px"
                        placeholder="Cancel"
                        colorType="purple"
                        radiusType="circle"
                        backgroundColorType="none"
                        fontSize="14px"
                        fontWeight="bold"
                        aria-haspopup="true"
                        // onClick={() => setOpenFunctionForm(true)}
                    />
                    <Button
                        className="modal-btn"
                        width="90px"
                        height="32px"
                        placeholder="Add"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="14px"
                        fontWeight="bold"
                        aria-haspopup="true"
                        // onClick={() => setOpenFunctionForm(true)}
                    />
                </div>
            </DialogContent>
        </Dialog>
    );
}
export default HubMarketplace;
