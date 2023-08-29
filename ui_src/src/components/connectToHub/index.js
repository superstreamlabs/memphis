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

import DialogContent from '@material-ui/core/DialogContent';
import { makeStyles } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import React, { useState } from 'react';

import { ReactComponent as CloseIcon } from '../../assets/images/close.svg';
import Input from '../Input';
import Button from '../button';
import CheckboxComponent from '../checkBox';

const useStyles = makeStyles((theme) => ({
    root: {
        flexGrow: 1
    },
    dialogPaper: {
        height: '50vh',
        minHeight: '550px',
        width: '25vw',
        minWidth: '440px',
        borderRadius: '10px',
        padding: '15px'
    }
}));

function ConnectToHub(props) {
    const classes = useStyles();
    const [formFields, setFormFields] = useState({
        username: '',
        password: '',
        rememberMe: true
    });

    const handelChangeUsername = (e) => {
        setFormFields({ ...formFields, username: e.target.value });
    };

    const handelChangePassword = (e) => {
        setFormFields({ ...formFields, password: e.target.value });
    };

    const handelChangeRememberMe = () => {
        setFormFields({ ...formFields, rememberMe: !formFields.rememberMe });
    };

    const clearFormAndClose = () => {
        setFormFields({
            username: '',
            password: '',
            rememberMe: true
        });
        props.closeModeal(false);
    };

    return (
        <Dialog
            open={props.open}
            onClose={(_, reson) => {
                if (reson === 'backdropClick') clearFormAndClose();
                // { props.clickOutside() }
            }}
            classes={{ paper: classes.dialogPaper }}
        >
            <DialogContent className={classes.dialogContent}>
                <div className="connect-to-hub">
                    <div className="connect-to-hub-header">
                        <p>Sign in to hub</p>
                        <CloseIcon alt="close" onClick={clearFormAndClose} width={12} height={12} />
                    </div>
                    <div className="user-password-sectoin">
                        <div className="user-name-input">
                            <p>Username</p>
                            <Input
                                value={formFields.username}
                                placeholder="Type usernmane"
                                type="text"
                                radiusType="semi-round"
                                borderColorType="none"
                                boxShadowsType="gray"
                                colorType="black"
                                backgroundColorType="none"
                                width="21vw"
                                minWidth="360px"
                                height="40px"
                                iconComponent=""
                                onChange={(e) => handelChangeUsername(e)}
                            />
                        </div>
                        <div className="password-input">
                            <p>Password</p>
                            <Input
                                value={formFields.password}
                                placeholder="Type password"
                                type="text"
                                radiusType="semi-round"
                                borderColorType="none"
                                boxShadowsType="gray"
                                colorType="black"
                                backgroundColorType="none"
                                width="21vw"
                                minWidth="360px"
                                height="40px"
                                iconComponent=""
                                onChange={(e) => handelChangePassword(e)}
                            />
                        </div>

                        <span className="remember-me-checkbox" onClick={handelChangeRememberMe}>
                            <CheckboxComponent checked={formFields.rememberMe} id={'checkedG'} onChange={handelChangeRememberMe} name={'checkedG'} />
                            <p>Remember me</p>
                        </span>
                    </div>
                    <div className="sign-in-btn">
                        <Button
                            className="modal-btn"
                            width="100%"
                            height="40px"
                            placeholder="Sign in"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="14px"
                            fontWeight="bold"
                            onClick={() => {}}
                        />
                        <p>Forgot password?</p>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}
export default ConnectToHub;
